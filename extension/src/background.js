const API_BASE = "http://localhost:8080/api/v1";

chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
  if (request.type === "FETCH_SPEEDS") {
    fetchSpeeds(request.bbox)
      .then((data) => sendResponse({ success: true, data }))
      .catch((err) => sendResponse({ success: false, error: err.message }));
    return true;
  }
});

async function fetchSpeeds(bbox) {
  const url = `${API_BASE}/speeds?bbox=${bbox.minLat},${bbox.minLng},${bbox.maxLat},${bbox.maxLng}`;
  const response = await fetch(url);
  if (!response.ok) {
    throw new Error(`API error: ${response.status}`);
  }
  return response.json();
}
