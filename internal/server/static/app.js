const RETRIABLE = [502, 503, 504];

const screens = {
  input: document.getElementById("screen-input"),
  clarify: document.getElementById("screen-clarify"),
  summary: document.getElementById("screen-summary"),
  success: document.getElementById("screen-success"),
  loading: document.getElementById("screen-loading"),
};

const errorBox = document.getElementById("error");
const loadingLabel = document.getElementById("loading-label");
const inputText = document.getElementById("input-text");
const inputAnswer = document.getElementById("input-answer");
const clarifyQuestion = document.getElementById("clarify-question");

let conversation = [];
let pendingEvent = null;
let currentScreen = "input";

function show(name, label) {
  if (name !== "loading") currentScreen = name;
  for (const [key, el] of Object.entries(screens)) {
    el.hidden = key !== name;
  }
  if (name === "loading") loadingLabel.textContent = label;
  if (name === "input") inputText.focus();
  if (name === "clarify") inputAnswer.focus();
}

function showError(message) {
  errorBox.textContent = message;
  errorBox.hidden = false;
}

function clearError() {
  errorBox.hidden = true;
}

function friendlyError(err) {
  console.error(err);
  if (err.status === 400) return err.message;
  if (err.status === 500) return "Server problem - check that GEMINI_API_KEY is set.";
  if (RETRIABLE.includes(err.status)) return "The model is busy right now. Try again in a moment.";
  return "Something went wrong. Try again.";
}

// retries defaults to 0 
async function postJSON(url, body, retries = 0) {
  for (let attempt = 0; ; attempt++) {
    const res = await fetch(url, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });
    const data = await res.json().catch(() => ({}));
    if (res.ok) return data;

    if (RETRIABLE.includes(res.status) && attempt < retries) {
      show("loading", "Model is busy, retrying...");
      await new Promise((r) => setTimeout(r, 800 * (attempt + 1)));
      continue;
    }

    const err = new Error(data.error || `request failed (${res.status})`);
    err.status = res.status;
    throw err;
  }
}

function formatDate(value) {
  const [y, m, d] = (value || "").split("-").map(Number);
  if (!y || !m || !d) return value;
  return new Date(y, m - 1, d).toLocaleDateString(undefined, {
    weekday: "short",
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

function formatTime(value) {
  const [h, m] = (value || "").split(":").map(Number);
  if (Number.isNaN(h) || Number.isNaN(m)) return value;
  const date = new Date();
  date.setHours(h, m, 0, 0);
  return date.toLocaleTimeString(undefined, { hour: "numeric", minute: "2-digit" });
}

function formatDuration(minutes) {
  if (!minutes) return "-";
  const h = Math.floor(minutes / 60);
  const m = minutes % 60;
  if (!h) return `${m} min`;
  return m ? `${h}h ${m}m` : `${h}h`;
}

function showSummary(event) {
  pendingEvent = event;
  document.getElementById("summary-title").textContent = event.title || "(untitled)";
  document.getElementById("summary-date").textContent = formatDate(event.date);
  document.getElementById("summary-time").textContent = formatTime(event.start_time);
  document.getElementById("summary-duration").textContent = formatDuration(event.duration_minutes);
  show("summary");
}

async function extract(candidate) {
  clearError();
  show("loading", "Reading your event...");
  try {
    const event = await postJSON("/api/extract", { conversation: candidate }, 2);
    conversation = candidate;

    if (event.needs_clarification) {
      clarifyQuestion.textContent = event.clarifying_question;
      pendingEvent = event;
      inputAnswer.value = "";
      show("clarify");
      return;
    }

    showSummary(event);
  } catch (err) {
    showError(friendlyError(err));
    show(currentScreen);
  }
}

function reset() {
  conversation = [];
  pendingEvent = null;
  inputText.value = "";
  inputAnswer.value = "";
  clearError();
  show("input");
}

document.getElementById("form-input").addEventListener("submit", (e) => {
  e.preventDefault();
  const text = inputText.value.trim();
  if (!text) return;
  extract([text]);
});

document.getElementById("form-clarify").addEventListener("submit", (e) => {
  e.preventDefault();
  const answer = inputAnswer.value.trim();
  if (!answer) return;
  extract([...conversation, `Q: ${clarifyQuestion.textContent}`, `A: ${answer}`]);
});


document.getElementById("btn-create").addEventListener("click", async () => {
  clearError();
  show("loading", "Adding to your calendar...");
  try {
    const { link } = await postJSON("/api/create", {
      title: pendingEvent.title,
      date: pendingEvent.date,
      start_time: pendingEvent.start_time,
      duration_minutes: pendingEvent.duration_minutes,
    });
    document.getElementById("success-link").href = link;
    show("success");
  } catch (err) {
    showError(friendlyError(err));
    show("summary");
  }
});

document.getElementById("btn-cancel").addEventListener("click", reset);
document.getElementById("btn-another").addEventListener("click", reset);

show("input");
