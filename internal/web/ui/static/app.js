// Gallery front end. Everything user-supplied (filenames, tags, URLs) reaches
// the DOM through textContent or attributes, never through HTML strings.
(() => {
  "use strict";

  const PAGE_SIZE = 48;
  const POLL_MS = 2500;
  const MAX_FILE_BYTES = 10 << 20;
  const UPLOAD_CONCURRENCY = 3;
  const ACCEPTED_TYPES = ["image/jpeg", "image/png", "image/gif", "image/webp"];
  const GOOGLE_FONTS = { Roboto: "Roboto", "Open Sans": "Open+Sans", Lato: "Lato", Montserrat: "Montserrat" };
  const FAULT_FIELDS = ["latency_ms", "latency_probability", "error_probability", "slow_db_ms", "worker_failure_probability", "worker_delay_ms"];

  const $ = (id) => document.getElementById(id);
  const ui = {
    count: $("count"), filters: $("filters"), tagRail: $("tagRail"), matchMode: $("matchMode"),
    clearFilters: $("clearFilters"), slides: $("slides"), empty: $("empty"), loadState: $("loadState"),
    sentinel: $("sentinel"), backdrop: $("backdrop"), dropOverlay: $("dropOverlay"), toasts: $("toasts"),
    faultsBadge: $("faultsBadge"), slideTpl: $("slideTpl"),
  };

  const state = {
    images: [],
    nodes: new Map(), // image id -> { sig, li }
    page: 0,
    totalPages: 1,
    total: 0,
    loading: false,
    loadError: null,
    generation: 0, // bumped on every filter change so stale responses are dropped
    tags: [],
    matchAll: false,
    tagColors: {},
    predefined: [],
    settings: null,
    viewing: null,
    pollTimer: 0,
    animateNew: false, // set by refresh() so slides that arrive later settle onto the table
  };

  // ---------- helpers ----------

  function el(tag, attrs = {}, ...children) {
    const node = document.createElement(tag);
    for (const [k, v] of Object.entries(attrs)) {
      if (v == null || (v === false && !k.startsWith("aria-"))) continue;
      if (k === "class") node.className = v;
      else if (k.startsWith("on")) node.addEventListener(k.slice(2), v);
      else node.setAttribute(k, v === true && !k.startsWith("aria-") ? "" : String(v));
    }
    node.append(...children.flat().filter((c) => c != null && c !== false));
    return node;
  }

  async function api(path, { method = "GET", json } = {}) {
    const init = { method, headers: { Accept: "application/json" } };
    if (json !== undefined) {
      init.headers["Content-Type"] = "application/json";
      init.body = JSON.stringify(json);
    }
    const res = await fetch(path, init);
    if (!res.ok) {
      const text = (await res.text()).trim();
      const err = new Error(text || `${res.status} ${res.statusText}`);
      err.status = res.status;
      throw err;
    }
    return res.json();
  }

  function formatBytes(n) {
    if (!n) return "0 B";
    const units = ["B", "KB", "MB", "GB"];
    const i = Math.min(Math.floor(Math.log(n) / Math.log(1024)), units.length - 1);
    const v = n / 1024 ** i;
    return `${v >= 10 || i === 0 ? Math.round(v) : v.toFixed(1)} ${units[i]}`;
  }

  function formatLabel(contentType) {
    return { "image/jpeg": "JPEG", "image/jpg": "JPEG", "image/png": "PNG", "image/gif": "GIF", "image/webp": "WebP" }[contentType] || "";
  }

  const dateFmt = new Intl.DateTimeFormat(undefined, { dateStyle: "medium", timeStyle: "short" });
  const plural = (n, one, many) => `${n.toLocaleString()} ${n === 1 ? one : many}`;
  const isDeveloping = (img) => img.status === "pending" || img.status === "processing";
  const tagColor = (name) => state.tagColors[name] || predefinedColor(name);
  const predefinedColor = (name) => (state.predefined.find((t) => t.name === name) || {}).color;

  function toast(message, { error = false } = {}) {
    const node = el("div", { class: error ? "toast is-error" : "toast", role: error ? "alert" : "status" }, message);
    ui.toasts.append(node);
    // Re-showing moves the popover above whichever dialog opened last.
    if (ui.toasts.matches(":popover-open")) ui.toasts.hidePopover();
    ui.toasts.showPopover();
    setTimeout(() => {
      node.remove();
      if (!ui.toasts.childElementCount && ui.toasts.matches(":popover-open")) ui.toasts.hidePopover();
    }, error ? 7000 : 4000);
  }

  function lightDismiss(dialog) {
    dialog.addEventListener("click", (e) => { if (e.target === dialog) dialog.close(); });
    dialog.querySelectorAll("[data-close]").forEach((b) => b.addEventListener("click", () => dialog.close()));
  }

  function confirmAction({ title, body, ok }) {
    const dialog = $("confirmDialog");
    $("confirmTitle").textContent = title;
    $("confirmBody").textContent = body;
    $("confirmOk").textContent = ok;
    dialog.returnValue = "";
    dialog.showModal();
    return new Promise((resolve) => dialog.addEventListener("close", () => resolve(dialog.returnValue === "ok"), { once: true }));
  }

  const typing = (e) => e.target.closest("input, textarea, select, [contenteditable]");

  // ---------- listing ----------

  function listQuery(page, pageSize) {
    const p = new URLSearchParams({ page: String(page), page_size: String(pageSize) });
    state.tags.forEach((t) => p.append("tags", t));
    if (state.matchAll && state.tags.length > 1) p.set("match_all", "true");
    return p;
  }

  function absorb(data) {
    Object.assign(state.tagColors, data.tag_colors || {});
    state.total = data.total_count;
    state.totalPages = data.total_pages;
  }

  async function loadNextPage() {
    if (state.loading || state.page >= state.totalPages) return;
    const gen = state.generation;
    state.loading = true;
    state.loadError = null;
    renderLoadState();
    try {
      const data = await api("/api/images?" + listQuery(state.page + 1, PAGE_SIZE));
      if (gen !== state.generation) return;
      absorb(data);
      state.page = data.page;
      const known = new Set(state.images.map((i) => i.id));
      state.images.push(...data.images.filter((i) => !known.has(i.id)));
    } catch (err) {
      if (gen === state.generation) state.loadError = err;
    } finally {
      if (gen === state.generation) {
        state.loading = false;
        render();
      }
    }
  }

  // Re-reads the top of the listing: picks up new uploads and status changes.
  async function refresh() {
    const gen = state.generation;
    const size = Math.min(Math.max(state.images.length, PAGE_SIZE), 100);
    try {
      const data = await api("/api/images?" + listQuery(1, size));
      if (gen !== state.generation) return;
      absorb(data);
      const fresh = new Set(data.images.map((i) => i.id));
      state.images = data.images.concat(state.images.filter((i) => !fresh.has(i.id)).slice(0, Math.max(0, state.total - data.images.length)));
      state.page = Math.max(1, Math.floor(state.images.length / PAGE_SIZE));
      state.animateNew = true;
      render();
    } catch (err) {
      // A failed background refresh keeps what is on screen; the next poll retries.
    }
  }

  function restart() {
    state.generation++;
    state.images = [];
    state.page = 0;
    state.totalPages = 1;
    state.loading = false;
    syncURL();
    render();
    loadNextPage();
  }

  function schedulePoll() {
    clearTimeout(state.pollTimer);
    if (state.images.some(isDeveloping) && !document.hidden) {
      state.pollTimer = setTimeout(refresh, POLL_MS);
    }
  }

  // ---------- rendering ----------

  function render() {
    renderSlides();
    renderCount();
    renderFilters();
    renderEmpty();
    renderLoadState();
    if (state.viewing && $("viewer").open) renderViewer();
    schedulePoll();
  }

  function slideSignature(img) {
    const s = state.settings || {};
    return [img.status, img.width, img.height, (img.tags || []).join(","), state.tags.join(","), s.show_tags, s.show_dimensions, s.show_content_type].join("|");
  }

  function renderSlides() {
    const seen = new Set();
    const nodes = state.images.map((img) => {
      seen.add(img.id);
      const sig = slideSignature(img);
      const cached = state.nodes.get(img.id);
      if (cached && cached.sig === sig) return cached.li;
      const li = slideNode(img);
      if (!cached && state.animateNew) li.classList.add("is-new");
      state.nodes.set(img.id, { sig, li });
      return li;
    });
    for (const id of state.nodes.keys()) if (!seen.has(id)) state.nodes.delete(id);
    ui.slides.replaceChildren(...nodes);
    state.animateNew = false;
  }

  function slideNode(img) {
    const s = state.settings || {};
    const li = ui.slideTpl.content.firstElementChild.cloneNode(true);
    li.dataset.status = img.status || "ready";
    li.dataset.id = img.id;
    li.querySelector(".frame-no").textContent = `No. ${img.id}`;

    const win = li.querySelector(".slide-window");
    win.setAttribute("aria-label", `View ${img.name || "image"}`);
    win.addEventListener("click", () => openViewer(img.id));

    const pic = li.querySelector("img");
    pic.src = `${img.thumbnail_url}?s=${encodeURIComponent(img.status || "ready")}`;
    pic.addEventListener("error", () => li.classList.add("is-broken"), { once: true });

    const name = li.querySelector(".slide-name");
    name.textContent = img.name || `Image ${img.id}`;
    name.title = img.name || "";

    const meta = li.querySelector(".slide-meta");
    if (isDeveloping(img)) meta.append(el("span", { class: "status status-developing" }, "Developing"));
    if (img.status === "failed") meta.append(el("span", { class: "status status-failed", title: img.processing_error || "" }, "Not processed"));
    if (s.show_dimensions !== false && img.width && img.height) meta.append(el("span", {}, `${img.width} × ${img.height}`));
    if (s.show_content_type !== false && formatLabel(img.content_type)) meta.append(el("span", {}, formatLabel(img.content_type)));
    meta.append(el("span", {}, formatBytes(img.size)));

    if (s.show_tags !== false) li.querySelector(".slide-tags").append(...(img.tags || []).map((t) => el("li", {}, pencilTag(t))));
    return li;
  }

  function pencilTag(name) {
    const btn = el("button", {
      class: "pencil-tag", type: "button", "aria-pressed": state.tags.includes(name),
      title: state.tags.includes(name) ? `Stop filtering by ${name}` : `Show only images tagged ${name}`,
      onclick: (e) => { e.stopPropagation(); toggleTag(name); },
    }, name);
    const color = tagColor(name);
    if (color) btn.style.setProperty("--tag", color);
    return btn;
  }

  function renderCount() {
    if (state.page === 0 && state.loading) { ui.count.textContent = ""; return; }
    if (!state.tags.length) { ui.count.textContent = plural(state.total, "image", "images"); return; }
    ui.count.textContent = `${plural(state.total, "image", "images")} ${state.total === 1 ? "matches" : "match"}`;
  }

  function renderFilters() {
    const names = new Set(state.tags);
    state.images.forEach((i) => (i.tags || []).forEach((t) => names.add(t)));
    const sorted = [...names].sort((a, b) => (state.tags.includes(b) - state.tags.includes(a)) || a.localeCompare(b));

    ui.filters.hidden = sorted.length === 0;
    ui.tagRail.replaceChildren(...sorted.map((name) => {
      const chip = el("button", {
        class: "tag-chip", type: "button", "aria-pressed": state.tags.includes(name), onclick: () => toggleTag(name),
      }, name);
      const color = tagColor(name);
      if (color) chip.style.setProperty("--tag", color);
      return chip;
    }));

    ui.matchMode.hidden = state.tags.length < 2;
    ui.matchMode.querySelectorAll("button").forEach((b) => b.setAttribute("aria-pressed", String((b.dataset.match === "all") === state.matchAll)));
    ui.clearFilters.hidden = state.tags.length === 0;
  }

  function renderEmpty() {
    const show = !state.loading && !state.loadError && state.images.length === 0 && state.page > 0;
    ui.empty.hidden = !show;
    if (!show) return;
    const action = $("emptyAction");
    if (state.tags.length) {
      $("emptyTitle").textContent = state.matchAll ? "No image has all of these tags" : "No images with these tags";
      $("emptyBody").textContent = `Nothing is tagged ${state.tags.join(state.matchAll ? " and " : " or ")} yet.`;
      action.textContent = "Clear filters";
      action.onclick = () => setFilters([], false);
    } else {
      $("emptyTitle").textContent = "No images yet";
      $("emptyBody").textContent = "Drop images anywhere on this page, or choose them from your computer.";
      action.textContent = "Upload images";
      action.onclick = () => openUpload();
    }
  }

  function renderLoadState() {
    ui.loadState.replaceChildren();
    if (state.loadError) {
      ui.loadState.append(
        el("span", {}, `Couldn't load images: ${state.loadError.message}`),
        el("button", { class: "btn btn-quiet", type: "button", onclick: loadNextPage }, "Try again"),
      );
    } else if (state.loading) {
      ui.loadState.append(el("span", {}, state.page === 0 ? "Loading images…" : "Loading more…"));
    } else if (state.page < state.totalPages && state.images.length) {
      ui.loadState.append(el("button", { class: "btn btn-quiet", type: "button", onclick: loadNextPage }, "Load more"));
    }
  }

  // ---------- filters ----------

  function setFilters(tags, matchAll) {
    state.tags = tags;
    state.matchAll = matchAll;
    restart();
  }

  function toggleTag(name) {
    const tags = state.tags.includes(name) ? state.tags.filter((t) => t !== name) : [...state.tags, name];
    setFilters(tags, state.matchAll);
  }

  function syncURL() {
    const p = new URLSearchParams();
    state.tags.forEach((t) => p.append("tags", t));
    if (state.matchAll && state.tags.length > 1) p.set("match_all", "true");
    const qs = p.toString();
    history.replaceState(null, "", "/gallery" + (qs ? "?" + qs : ""));
  }

  ui.matchMode.addEventListener("click", (e) => {
    const b = e.target.closest("button[data-match]");
    if (b) setFilters(state.tags, b.dataset.match === "all");
  });
  ui.clearFilters.addEventListener("click", () => setFilters([], false));

  // ---------- viewer ----------

  const viewer = $("viewer");
  lightDismiss(viewer);

  function openViewer(id) {
    state.viewing = id;
    renderViewer();
    if (!viewer.open) viewer.showModal();
  }

  function renderViewer() {
    const idx = state.images.findIndex((i) => i.id === state.viewing);
    if (idx < 0) { viewer.close(); return; }
    const img = state.images[idx];
    const pic = $("viewerImg");
    if (pic.dataset.id !== img.id) {
      pic.dataset.id = img.id;
      pic.src = img.url;
    }
    pic.alt = img.name || "";
    $("viewerFrame").textContent = `No. ${img.id}`;
    $("viewerName").textContent = img.name || `Image ${img.id}`;

    const status = $("viewerStatus");
    status.hidden = !(isDeveloping(img) || img.status === "failed");
    status.className = "viewer-status status " + (img.status === "failed" ? "status-failed" : "status-developing");
    status.textContent = img.status === "failed"
      ? `Processing failed${img.processing_error ? ": " + img.processing_error : ""}`
      : "Developing: the thumbnail and dimensions will appear shortly";

    const facts = [
      ["Size", formatBytes(img.size)],
      ["Dimensions", img.width && img.height ? `${img.width} × ${img.height} px` : null],
      ["Format", formatLabel(img.content_type) || null],
      ["Uploaded", img.uploaded_at ? dateFmt.format(new Date(img.uploaded_at)) : img.upload_time],
    ].filter(([, v]) => v);
    $("viewerFacts").replaceChildren(...facts.flatMap(([k, v]) => [el("dt", {}, k), el("dd", {}, v)]));
    $("viewerTags").replaceChildren(...(img.tags || []).map((t) => el("li", {}, pencilTag(t))));
    $("viewerOriginal").href = img.url;

    $("viewerPrev").disabled = idx === 0;
    $("viewerNext").disabled = idx === state.images.length - 1 && state.page >= state.totalPages;
    if (idx >= state.images.length - 3) loadNextPage();
  }

  function stepViewer(delta) {
    const idx = state.images.findIndex((i) => i.id === state.viewing);
    const next = state.images[idx + delta];
    if (next) openViewer(next.id);
  }

  $("viewerPrev").addEventListener("click", () => stepViewer(-1));
  $("viewerNext").addEventListener("click", () => stepViewer(1));
  viewer.addEventListener("keydown", (e) => {
    if (e.key === "ArrowLeft") stepViewer(-1);
    else if (e.key === "ArrowRight") stepViewer(1);
    else if (e.key === "Delete") deleteViewing();
  });
  viewer.addEventListener("close", () => { state.viewing = null; });
  $("viewerDelete").addEventListener("click", deleteViewing);

  async function deleteViewing() {
    const img = state.images.find((i) => i.id === state.viewing);
    if (!img) return;
    const name = img.name || `image ${img.id}`;
    const ok = await confirmAction({
      title: `Delete ${name}?`,
      body: "This removes the image and its thumbnail for everyone. You can't undo it.",
      ok: "Delete image",
    });
    if (!ok) return;
    try {
      await api(`/api/images/${encodeURIComponent(img.id)}`, { method: "DELETE" });
    } catch (err) {
      toast(`Couldn't delete ${name}: ${err.message}`, { error: true });
      return;
    }
    const idx = state.images.indexOf(img);
    state.images.splice(idx, 1);
    state.total = Math.max(0, state.total - 1);
    const neighbour = state.images[idx] || state.images[idx - 1];
    if (neighbour) state.viewing = neighbour.id; else viewer.close();
    render();
    toast(`Deleted ${name}`);
  }

  // ---------- upload ----------

  const uploadDialog = $("uploadDialog");
  const queue = [];
  const picked = new Set();
  let uploading = false;
  lightDismiss(uploadDialog);

  function openUpload(files) {
    if (!uploadDialog.open) {
      renderTagPicker();
      uploadDialog.showModal();
    }
    if (files && files.length) addFiles(files);
  }

  function addFiles(files) {
    for (const file of files) {
      const item = { file, status: "queued", progress: 0, error: "", preview: URL.createObjectURL(file) };
      if (!ACCEPTED_TYPES.includes(file.type)) Object.assign(item, { status: "rejected", error: "Not a JPEG, PNG, GIF or WebP image" });
      else if (file.size > MAX_FILE_BYTES) Object.assign(item, { status: "rejected", error: `${formatBytes(file.size)} is over the 10 MB limit` });
      queue.push(item);
    }
    renderQueue();
  }

  function renderQueue() {
    $("queue").replaceChildren(...queue.map((item) => {
      const stateText = {
        queued: formatBytes(item.file.size),
        uploading: `Uploading ${Math.round(item.progress * 100)}%`,
        done: "Uploaded",
        failed: item.error,
        rejected: item.error,
      }[item.status];
      const cls = item.status === "failed" || item.status === "rejected" ? "queue-state is-error" : item.status === "done" ? "queue-state is-done" : "queue-state";
      return el("li", {},
        el("img", { src: item.preview, alt: "" }),
        el("div", {},
          el("p", { class: "queue-name", title: item.file.name }, item.file.name),
          el("p", { class: cls }, stateText),
          item.status === "uploading" ? el("progress", { max: "1", value: String(item.progress) }) : null),
        item.status === "uploading" || item.status === "done" ? el("span") :
          el("button", { class: "icon-btn", type: "button", "aria-label": `Remove ${item.file.name}`, disabled: uploading, onclick: () => removeItem(item) },
            closeIcon()));
    }));
    const ready = queue.filter((i) => i.status === "queued" || i.status === "failed").length;
    const submit = $("uploadSubmit");
    submit.disabled = uploading || ready === 0;
    const retry = ready > 0 && queue.every((i) => i.status !== "queued");
    submit.textContent = uploading ? "Uploading…" : !ready ? "Upload" : `${retry ? "Retry" : "Upload"} ${plural(ready, "image", "images")}`;
    const rejected = queue.filter((i) => i.status === "rejected").length;
    $("uploadSummary").textContent = rejected ? `${plural(rejected, "file", "files")} can't be uploaded` : "";
  }

  function closeIcon() {
    const svg = document.createElementNS("http://www.w3.org/2000/svg", "svg");
    svg.setAttribute("class", "icon");
    svg.setAttribute("viewBox", "0 0 24 24");
    svg.setAttribute("aria-hidden", "true");
    const path = document.createElementNS("http://www.w3.org/2000/svg", "path");
    path.setAttribute("d", "M6 6l12 12M18 6L6 18");
    svg.append(path);
    return svg;
  }

  function removeItem(item) {
    URL.revokeObjectURL(item.preview);
    queue.splice(queue.indexOf(item), 1);
    renderQueue();
  }

  function renderTagPicker() {
    $("tagPicker").replaceChildren(...state.predefined.map((t) => {
      const chip = el("button", {
        class: "tag-chip", type: "button", "aria-pressed": picked.has(t.name), title: t.description || null,
        onclick: () => {
          if (picked.has(t.name)) picked.delete(t.name); else picked.add(t.name);
          chip.setAttribute("aria-pressed", String(picked.has(t.name)));
        },
      }, t.name);
      if (t.color) chip.style.setProperty("--tag", t.color);
      return chip;
    }));
  }

  function chosenTags() {
    const custom = $("customTags").value.split(",").map((t) => t.trim()).filter(Boolean);
    return [...new Set([...picked, ...custom])];
  }

  function uploadOne(item, tags) {
    return new Promise((resolve) => {
      const form = new FormData();
      form.append("files", item.file, item.file.name);
      if (tags.length) form.append("tags", tags.join(","));
      const xhr = new XMLHttpRequest();
      xhr.open("POST", "/api/images");
      xhr.setRequestHeader("Accept", "application/json");
      xhr.upload.addEventListener("progress", (e) => {
        if (e.lengthComputable) { item.progress = e.loaded / e.total; renderQueue(); }
      });
      xhr.addEventListener("load", () => {
        let body = null;
        try { body = JSON.parse(xhr.responseText); } catch { /* plain-text error body */ }
        if (xhr.status >= 200 && xhr.status < 300 && body && body.count > 0) {
          item.status = "done";
        } else {
          item.status = "failed";
          item.error = (body && body.errors && body.errors[0] && body.errors[0].error) || xhr.responseText.trim() || `Upload failed (HTTP ${xhr.status})`;
        }
        resolve();
      });
      xhr.addEventListener("error", () => {
        item.status = "failed";
        item.error = "The server didn't respond. Check your connection and retry.";
        resolve();
      });
      item.status = "uploading";
      item.progress = 0;
      renderQueue();
      xhr.send(form);
    });
  }

  $("uploadForm").addEventListener("submit", async (e) => {
    e.preventDefault();
    if (uploading) return;
    const batch = queue.filter((i) => i.status === "queued" || i.status === "failed");
    if (!batch.length) return;
    uploading = true;
    const tags = chosenTags();
    const pending = batch.slice();
    const worker = async () => { while (pending.length) await uploadOne(pending.shift(), tags); };
    await Promise.all(Array.from({ length: Math.min(UPLOAD_CONCURRENCY, batch.length) }, worker));
    uploading = false;
    renderQueue();

    const done = batch.filter((i) => i.status === "done").length;
    const failed = batch.length - done;
    if (done) {
      toast(failed ? `Uploaded ${plural(done, "image", "images")}; ${failed} failed` : `Uploaded ${plural(done, "image", "images")}`, { error: false });
      if (state.tags.length) setFilters([], false); else refresh();
    } else {
      toast("Nothing was uploaded. See the errors in the list.", { error: true });
    }
    if (!failed && !queue.some((i) => i.status === "rejected")) uploadDialog.close();
  });

  uploadDialog.addEventListener("cancel", (e) => { if (uploading) e.preventDefault(); });
  uploadDialog.addEventListener("close", () => {
    if (uploading) { uploadDialog.showModal(); return; }
    queue.splice(0).forEach((i) => URL.revokeObjectURL(i.preview));
    picked.clear();
    $("customTags").value = "";
    $("fileInput").value = "";
    renderQueue();
  });

  $("fileInput").addEventListener("change", (e) => { addFiles(e.target.files); e.target.value = ""; });
  $("openUpload").addEventListener("click", () => openUpload());

  const dropzone = $("dropzone");
  ["dragenter", "dragover"].forEach((t) => dropzone.addEventListener(t, () => dropzone.classList.add("is-over")));
  ["dragleave", "drop"].forEach((t) => dropzone.addEventListener(t, () => dropzone.classList.remove("is-over")));

  let dragDepth = 0;
  const hasFiles = (e) => e.dataTransfer && [...e.dataTransfer.types].includes("Files");
  window.addEventListener("dragenter", (e) => {
    if (!hasFiles(e)) return;
    dragDepth++;
    if (!uploadDialog.open) ui.dropOverlay.hidden = false;
  });
  window.addEventListener("dragleave", () => { if (--dragDepth <= 0) { dragDepth = 0; ui.dropOverlay.hidden = true; } });
  window.addEventListener("dragover", (e) => { if (hasFiles(e)) e.preventDefault(); });
  window.addEventListener("drop", (e) => {
    if (!hasFiles(e)) return;
    e.preventDefault();
    dragDepth = 0;
    ui.dropOverlay.hidden = true;
    openUpload(e.dataTransfer.files);
  });
  window.addEventListener("paste", (e) => {
    const files = [...(e.clipboardData ? e.clipboardData.files : [])].filter((f) => f.type.startsWith("image/"));
    if (files.length && !typing(e)) openUpload(files);
  });

  // ---------- settings ----------

  const settingsDialog = $("settingsDialog");
  const settingsForm = $("settingsForm");
  lightDismiss(settingsDialog);

  function applySettings(s) {
    const root = document.documentElement;
    root.dataset.theme = s.text_theme === "dark" ? "dark" : "light";
    root.style.setProperty("--cols", String(s.grid_columns || 5));

    const family = s.font_family || "system-ui";
    if (GOOGLE_FONTS[family] && !document.querySelector(`link[data-font="${GOOGLE_FONTS[family]}"]`)) {
      document.head.append(el("link", { rel: "stylesheet", "data-font": GOOGLE_FONTS[family], href: `https://fonts.googleapis.com/css2?family=${GOOGLE_FONTS[family]}:wght@400;600&display=swap` }));
    }
    root.style.setProperty("--font-body", family === "system-ui" ? "system-ui, sans-serif" : `"${family}", system-ui, sans-serif`);

    const url = s.background_image_url || "";
    const bd = ui.backdrop.style;
    if (/^https?:\/\//i.test(url)) {
      bd.backgroundImage = `url("${url.replace(/["\\\n\r]/g, encodeURIComponent)}")`;
      bd.backgroundSize = s.background_style === "repeat" ? "auto" : s.background_style || "cover";
      bd.backgroundRepeat = s.background_style === "repeat" ? "repeat" : "no-repeat";
      bd.opacity = String(s.background_opacity ?? 0.3);
    } else {
      bd.backgroundImage = "";
      bd.opacity = "0";
    }
  }

  function fillSettingsForm(s) {
    const f = settingsForm.elements;
    f.text_theme.value = s.text_theme;
    f.font_family.value = s.font_family;
    f.grid_columns.value = s.grid_columns;
    f.show_tags.checked = s.show_tags;
    f.show_dimensions.checked = s.show_dimensions;
    f.show_content_type.checked = s.show_content_type;
    f.background_image_url.value = s.background_image_url || "";
    f.background_style.value = s.background_style;
    f.background_opacity.value = s.background_opacity;
    syncOutputs();
  }

  function readSettingsForm() {
    const f = settingsForm.elements;
    return {
      text_theme: f.text_theme.value,
      font_family: f.font_family.value,
      grid_columns: Number(f.grid_columns.value),
      show_tags: f.show_tags.checked,
      show_dimensions: f.show_dimensions.checked,
      show_content_type: f.show_content_type.checked,
      background_image_url: f.background_image_url.value.trim(),
      background_style: f.background_style.value,
      background_opacity: Number(f.background_opacity.value),
    };
  }

  function syncOutputs() {
    const f = settingsForm.elements;
    $("colsOut").textContent = f.grid_columns.value;
    $("opacityOut").textContent = `${Math.round(Number(f.background_opacity.value) * 100)}%`;
  }

  function preview() {
    syncOutputs();
    const draft = { ...state.settings, ...readSettingsForm() };
    applySettings(draft);
    const saved = state.settings;
    state.settings = draft;
    renderSlides();
    state.settings = saved;
  }

  settingsForm.addEventListener("input", preview);
  settingsForm.addEventListener("submit", async (e) => {
    e.preventDefault();
    const err = $("settingsError");
    err.textContent = "";
    try {
      state.settings = await api("/api/settings", { method: "PUT", json: readSettingsForm() });
    } catch (ex) {
      err.textContent = ex.message;
      return;
    }
    settingsDialog.close();
    toast("Settings saved");
  });
  settingsDialog.addEventListener("close", () => {
    // Closing without saving drops the live preview.
    applySettings(state.settings);
    render();
  });
  $("resetSettings").addEventListener("click", async () => {
    try {
      state.settings = await api("/api/settings/reset", { method: "POST" });
    } catch (ex) {
      $("settingsError").textContent = ex.message;
      return;
    }
    fillSettingsForm(state.settings);
    applySettings(state.settings);
    renderSlides();
    toast("Settings reset to defaults");
  });

  function openSettings(focusFaults) {
    if (state.settings) fillSettingsForm(state.settings);
    $("settingsError").textContent = "";
    settingsDialog.showModal();
    loadFaults();
    if (focusFaults) $("faultsForm").scrollIntoView({ block: "start" });
  }
  $("openSettings").addEventListener("click", () => openSettings(false));
  ui.faultsBadge.addEventListener("click", () => openSettings(true));

  // ---------- fault injection ----------

  const faultsForm = $("faultsForm");

  function faultsActive(c) {
    return c.error_probability > 0 || (c.latency_ms > 0 && c.latency_probability > 0) || c.slow_db_ms > 0 ||
      c.worker_failure_probability > 0 || c.worker_delay_ms > 0;
  }

  function showFaults(c) {
    const f = faultsForm.elements;
    FAULT_FIELDS.forEach((k) => { f[k].value = c[k] ?? 0; });
    f.latency_routes.value = (c.latency_routes || []).join(", ");
    ui.faultsBadge.hidden = !faultsActive(c);
  }

  async function loadFaults() {
    try {
      showFaults(await api("/api/settings/demo"));
      faultsForm.hidden = false;
    } catch {
      faultsForm.hidden = true; // 404 when DEMO_CONTROLS_ENABLED=false
    }
  }

  faultsForm.addEventListener("submit", async (e) => {
    e.preventDefault();
    const f = faultsForm.elements;
    const body = { latency_routes: f.latency_routes.value.split(",").map((s) => s.trim()).filter(Boolean) };
    FAULT_FIELDS.forEach((k) => { body[k] = Number(f[k].value || 0); });
    const status = $("faultsStatus");
    try {
      const c = await api("/api/settings/demo", { method: "PUT", json: body });
      showFaults(c);
      status.className = "hint";
      status.textContent = faultsActive(c) ? "Faults applied" : "No faults active";
    } catch (ex) {
      status.className = "hint is-error";
      status.textContent = ex.message;
    }
  });
  $("faultsOff").addEventListener("click", async () => {
    const status = $("faultsStatus");
    try {
      showFaults(await api("/api/settings/demo/reset", { method: "POST" }));
      status.className = "hint";
      status.textContent = "All faults off";
    } catch (ex) {
      status.className = "hint is-error";
      status.textContent = ex.message;
    }
  });

  // ---------- boot ----------

  document.addEventListener("keydown", (e) => {
    if (e.key === "u" && !e.metaKey && !e.ctrlKey && !e.altKey && !typing(e) && !document.querySelector("dialog[open]")) {
      e.preventDefault();
      openUpload();
    }
  });
  document.addEventListener("visibilitychange", schedulePoll);

  new IntersectionObserver((entries) => {
    if (entries.some((en) => en.isIntersecting) && state.page > 0 && !state.loadError) loadNextPage();
  }, { rootMargin: "800px 0px" }).observe(ui.sentinel);

  async function boot() {
    const params = new URLSearchParams(location.search);
    state.tags = params.getAll("tags");
    state.matchAll = params.get("match_all") === "true";

    const [settings, predefined] = await Promise.allSettled([api("/api/settings"), api("/api/tags/predefined")]);
    if (settings.status === "fulfilled") state.settings = settings.value;
    if (predefined.status === "fulfilled") state.predefined = predefined.value;
    if (state.settings) applySettings(state.settings);

    restart();
    api("/api/settings/demo").then(showFaults, () => {});
  }

  boot();
})();
