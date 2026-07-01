# /speed-data — Add or Edit Speed Limit Data

Add new speed limit segments or update existing ones in the GeoJSON database.
Every change requires online validation and results in a pull request.

## Arguments

- `/speed-data <description>` — describe the speed limit to add or update in plain language
- `/speed-data list` — show all GeoJSON data files and segment counts

**Examples:**
- `/speed-data Makenji to Montezuma is 50 km/h with active speed cameras`
- `/speed-data Update A109 Mlolongo to 50 km/h, it's a built-up area now`
- `/speed-data Add Kangundo Road from Cabanas to Tala, 50 km/h urban`
- `/speed-data The speed limit on Waiyaki Way was changed to 60 km/h`

## Workflow

### Step 1: Parse the Request

Extract from the user's description:
- **Road name** (e.g., "A2 — Makenji to Montezuma", "Kangundo Road")
- **Speed limit** in km/h
- **Location** (towns, junctions, landmarks)
- **Direction** (both, northbound, southbound, etc.)
- **Any enforcement info** (speed cameras, NTSA zones)
- **Whether this is an ADD (new segment) or EDIT (update existing)**

If any critical info is missing (speed limit, location), ask the user before proceeding.

### Step 2: Check Existing Data

Search the GeoJSON files for any existing data covering this road or area:

```bash
grep -rl -i "<road_keyword>" data/geojson/
```

If matches are found, read the relevant file and show the user what currently exists:
- Segment name, speed limit, coordinates, source, verified status
- Identify whether this is an update to existing data or a new segment to add

If no matches, identify the best file to add the new segment to based on:
- Highway segments → named highway file (e.g., `a2_thika_nanyuki.geojson`)
- Urban zones → county-based urban file (e.g., `central_kenya_urban.geojson`)
- If no suitable file exists, create a new one following the naming convention

### Step 3: Online Validation (REQUIRED)

**Every data change MUST be validated online before applying.** Search for evidence:

1. **Search for official sources:**
   - Web search: `"<road name>" speed limit Kenya NTSA`
   - Web search: `"<road name>" speed cameras Kenya`
   - Web search: `"<location>" Kenya gazette speed limit`

2. **Search for driver reports and news:**
   - Web search: `"<road name>" <location> speed fine Kenya`
   - Web search: `"<road name>" accident Kenya`

3. **Cross-reference with legal framework:**
   - Built-up areas (towns, trading centres) → 50 km/h (Traffic Act Cap 403 Section 42)
   - School zones, hospitals, playgrounds → 30 km/h
   - Single carriageway highways → 100 km/h (LN 62/1975)
   - Dual carriageway highways → 110 km/h (LN 62/1975)
   - Nairobi Expressway → 80 km/h (NTSA directive, special exception)
   - PSVs/commercial vehicles → 80 km/h on any road

4. **Determine coordinates:**
   - Search for the location on OpenStreetMap to get accurate coordinates
   - Web search: `"<town name>" Kenya coordinates` or `"<junction name>" GPS coordinates`
   - Coordinates must be `[longitude, latitude]` (GeoJSON standard)
   - Verify coordinates are in Kenya (latitude roughly -4.5 to 4.5, longitude roughly 33.5 to 42)

**Present a validation summary to the user:**
```
VALIDATION REPORT
─────────────────
Road: <name>
Segment: <from> → <to>
Speed Limit: <X> km/h
Legal Basis: <Traffic Act section or gazette notice>
Evidence Found:
  - <source 1: description + URL>
  - <source 2: description + URL>
  - <source 3: description + URL>
Coordinates: [lng, lat] → [lng, lat]
Confidence: HIGH / MEDIUM / LOW
```

- **HIGH confidence**: Official gazette notice, NTSA announcement, or multiple consistent driver reports
- **MEDIUM confidence**: Consistent with legal framework + some driver reports, but no direct gazette citation
- **LOW confidence**: Only user report, no corroboration found

If confidence is LOW, warn the user and ask if they want to proceed with `"verified": false`.

### Step 4: Apply the Change

Create a new git branch for the change:
```bash
git checkout -b data/<slugified-description>
```

**For NEW segments** — add a Feature to the appropriate GeoJSON file.
Every feature MUST have ALL required properties:

```json
{
  "type": "Feature",
  "properties": {
    "road_name": "<road name — format: 'A2 — Segment Description' for highways, plain name for urban>",
    "speed_limit_kmh": <number>,
    "road_class": "<urban|peri_urban|highway|expressway>",
    "direction": "<both|northbound|southbound|eastbound|westbound>",
    "source": "<legal reference or source description>",
    "verified": <true if HIGH/MEDIUM confidence, false if LOW>,
    "county": "<Kenya county name>",
    "last_updated": "<YYYY-MM-DD>"
  },
  "geometry": {
    "type": "LineString",
    "coordinates": [
      [<longitude>, <latitude>],
      [<longitude>, <latitude>]
    ]
  }
}
```

**For EDITS to existing segments:**
- Show the user the before/after diff
- Update `speed_limit_kmh`, `source`, `verified`, `last_updated`, and any other changed fields
- If a segment needs to be split (e.g., a 100 km/h highway segment now has a 50 km/h zone in the middle), split it into separate features with correct coordinates

**road_class rules:**
- `urban` — built-up areas, towns, trading centres (typically 50 km/h or 30 km/h)
- `peri_urban` — transition zones (51–80 km/h)
- `highway` — single or dual carriageway open road (81–110 km/h)
- `expressway` — controlled-access expressway (Nairobi Expressway only, 80 km/h)

### Step 5: Validate the GeoJSON

After editing, validate the JSON is well-formed:
```bash
python3 -c "import json; json.load(open('<file>'))" && echo "Valid JSON"
```

Check that coordinates are within Kenya bounds:
```bash
python3 -c "
import json
data = json.load(open('<file>'))
for f in data['features']:
    for c in f['geometry']['coordinates']:
        lng, lat = c
        assert 33 <= lng <= 42, f'Bad longitude: {lng}'
        assert -5 <= lat <= 5, f'Bad latitude: {lat}'
print('All coordinates valid')
"
```

### Step 6: Regenerate Static Fallback

```bash
python3 -c "
import json, glob, os
all_features = []
for f in sorted(glob.glob('data/geojson/*.geojson')):
    with open(f) as fh:
        data = json.load(fh)
        all_features.extend(data.get('features', []))
output = {'type': 'FeatureCollection', 'features': all_features}
with open('frontend/static/speeds.json', 'w') as out:
    json.dump(output, out, indent=2)
print(f'Written {len(all_features)} features to frontend/static/speeds.json')
"
```

### Step 7: Commit and Create PR

Stage and commit the changes:
```bash
git add data/geojson/<file>.geojson frontend/static/speeds.json
git commit -m "<descriptive message about the speed data change>"
```

Push and create a pull request:
```bash
git push -u origin data/<branch-name>
gh pr create --title "data: <short description>" --body "$(cat <<'EOF'
## Speed Data Update

**Road:** <road name>
**Segment:** <from> → <to>
**Speed Limit:** <X> km/h
**Road Class:** <class>
**County:** <county>

## Validation

**Confidence:** <HIGH/MEDIUM/LOW>
**Legal Basis:** <reference>

**Evidence:**
- <source 1>
- <source 2>

## Changes
- <what was added/modified>
- <number of segments affected>

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

### Step 8: Report

Output a summary:
```
DONE
────
Branch: data/<name>
PR: <URL>
File: data/geojson/<file>.geojson
Segments added/modified: <N>
Static fallback: regenerated (<total> features)
```

Remind the user to run a deploy after merge if they want it live on speed.koru.africa.

## Important Rules

- **NEVER skip online validation** — even if the user is confident, always verify
- **NEVER fabricate coordinates** — use OpenStreetMap or web search to find real coordinates
- **ALWAYS use [longitude, latitude] order** in GeoJSON (not [lat, lng])
- **ALWAYS regenerate frontend/static/speeds.json** after any GeoJSON change
- **ALWAYS create a PR** — never commit directly to main
- **Mark `verified: false`** if confidence is LOW — this flags it for future verification
- **Include enforcement info in the `source` field** — e.g., "NTSA speed cameras active"
- **Use today's date** for `last_updated`
