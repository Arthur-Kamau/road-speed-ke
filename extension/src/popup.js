const API_BASE = "http://localhost:8080";

async function checkStatus() {
  const dot = document.getElementById("statusDot");
  const text = document.getElementById("statusText");

  try {
    const res = await fetch(`${API_BASE}/health`);
    if (res.ok) {
      dot.classList.remove("offline");
      const stats = await fetch(`${API_BASE}/api/v1/stats`).then((r) => r.json());
      text.textContent = `Connected — ${stats.total_segments} segments`;
    } else {
      dot.classList.add("offline");
      text.textContent = "API error";
    }
  } catch {
    dot.classList.add("offline");
    text.textContent = "API offline — start the server";
  }
}

checkStatus();
