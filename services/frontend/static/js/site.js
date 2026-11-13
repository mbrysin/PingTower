"use strict";
const API_BASE = "http://localhost:8080";
const DAY_MS = 24 * 60 * 60 * 1000;
const WINDOW_DAYS = 30;
const DETAILED_HISTORY_DAYS = 90;
const DETAIL_BUCKET_MS = 5 * 60 * 1000;
const siteID = document.body.dataset.siteId;
if (!siteID) {
    throw new Error("Missing site ID on page");
}
async function loadSitePage(id) {
    const encodedID = encodeURIComponent(id);
    try {
        const [site, latestStat] = await Promise.all([
            fetchJSON(`/sites/${encodedID}`),
            fetchJSON(`/sites/${encodedID}/latest-stat`),
        ]);
        renderSite(site, latestStat);
    }
    catch (error) {
        console.error("Site page loading failed:", error);
        setText("#site-error", error instanceof Error ? error.message : String(error));
        return;
    }
    await loadAndDrawSelectedRange();
}
function setupDateControls() {
    const fromInput = getElement("#range-from");
    const toInput = getElement("#range-to");
    if (!fromInput || !toInput) {
        console.error("Date range controls are missing from the page");
        return;
    }
    const today = startOfDay(new Date());
    const firstDay = addDays(today, -(WINDOW_DAYS - 1));
    fromInput.value = toDateInputValue(firstDay);
    toInput.value = toDateInputValue(today);
    fromInput.addEventListener("change", () => {
        void loadAndDrawSelectedRange();
    });
    toInput.addEventListener("change", () => {
        void loadAndDrawSelectedRange();
    });
    getElement("#range-prev")?.addEventListener("click", () => shiftRange(-WINDOW_DAYS));
    getElement("#range-next")?.addEventListener("click", () => shiftRange(WINDOW_DAYS));
    getElement("#range-reset")?.addEventListener("click", () => {
        const currentDay = startOfDay(new Date());
        fromInput.value = toDateInputValue(addDays(currentDay, -(WINDOW_DAYS - 1)));
        toInput.value = toDateInputValue(currentDay);
        void loadAndDrawSelectedRange();
    });
}
function shiftRange(days) {
    const fromInput = getElement("#range-from");
    const toInput = getElement("#range-to");
    if (!fromInput || !toInput)
        return;
    const fromDate = parseDateInput(fromInput.value);
    const toDate = parseDateInput(toInput.value);
    if (!fromDate || !toDate)
        return;
    fromInput.value = toDateInputValue(addDays(fromDate, days));
    toInput.value = toDateInputValue(addDays(toDate, days));
    void loadAndDrawSelectedRange();
}
async function loadAndDrawSelectedRange() {
    const fromInput = getElement("#range-from");
    const toInput = getElement("#range-to");
    if (!fromInput || !toInput)
        return;
    const fromDate = parseDateInput(fromInput.value);
    const toDate = parseDateInput(toInput.value);
    if (!fromDate || !toDate) {
        setText("#chart-message", "Select a valid date range.");
        return;
    }
    if (fromDate.getTime() > toDate.getTime()) {
        setText("#chart-message", "The start date must not be after the end date.");
        return;
    }
    const detailedCutoff = startOfDay(new Date(Date.now() - DETAILED_HISTORY_DAYS * DAY_MS));
    const useDetailedStats = fromDate.getTime() >= detailedCutoff.getTime();
    const mode = getElement("#chart-mode");
    if (mode) {
        mode.textContent = useDetailedStats
            ? "Detailed checks — 5-minute availability"
            : "Daily availability";
    }
    setText("#chart-message", "Loading history…");
    try {
        const encodedID = encodeURIComponent(siteID);
        let points;
        if (useDetailedStats) {
            points = await loadDetailedPoints(encodedID, fromInput.value, toInput.value);
        }
        else {
            points = await loadDailyPoints(encodedID, fromInput.value, toInput.value);
        }
        drawChart(points, fromDate, toDate);
    }
    catch (error) {
        console.error("Chart loading failed:", error);
        setText("#chart-message", error instanceof Error
            ? `Could not load history: ${error.message}`
            : "Could not load availability history.");
    }
}
async function loadDetailedPoints(encodedID, fromValue, toValue) {
    const fromDate = parseDateInput(fromValue);
    const toDate = parseDateInput(toValue);
    if (!fromDate || !toDate)
        return [];
    const toExclusive = addDays(toDate, 1);
    const params = new URLSearchParams({
        from: fromDate.toISOString(),
        to: toExclusive.toISOString(),
    });
    const stats = await fetchJSON(`/sites/${encodedID}/stats?${params.toString()}`);
    return aggregateIntoFiveMinuteBuckets(stats, fromDate, toExclusive);
}
function aggregateIntoFiveMinuteBuckets(stats, from, toExclusive) {
    const buckets = new Map();
    for (const stat of stats) {
        const time = Date.parse(stat.checked_at);
        if (!Number.isFinite(time) || typeof stat.is_up !== "boolean") {
            continue;
        }
        const bucketTime = Math.floor(time / DETAIL_BUCKET_MS) * DETAIL_BUCKET_MS;
        const bucket = buckets.get(bucketTime) ?? {
            total: 0,
            successful: 0,
        };
        bucket.total += 1;
        if (stat.is_up)
            bucket.successful += 1;
        buckets.set(bucketTime, bucket);
    }
    const startTime = Math.floor(from.getTime() / DETAIL_BUCKET_MS) * DETAIL_BUCKET_MS;
    const endTime = toExclusive.getTime();
    const points = [];
    for (let time = startTime; time < endTime; time += DETAIL_BUCKET_MS) {
        const bucket = buckets.get(time);
        points.push({
            time,
            availability: bucket && bucket.total > 0
                ? (bucket.successful / bucket.total) * 100
                : null,
        });
    }
    return points;
}
async function loadDailyPoints(encodedID, fromValue, toValue) {
    const params = new URLSearchParams({
        from: fromValue,
        to: toValue,
    });
    const stats = await fetchJSON(`/sites/${encodedID}/daily-stats?${params.toString()}`);
    const pointsByDate = new Map();
    for (const stat of stats) {
        const dateKey = stat.stat_date.slice(0, 10);
        const date = parseDateInput(dateKey);
        if (!date)
            continue;
        let availability = stat.success_percent;
        if (availability == null &&
            Number.isFinite(stat.checks) &&
            stat.checks > 0) {
            availability =
                ((stat.checks - stat.failed_checks) / stat.checks) * 100;
        }
        pointsByDate.set(dateKey, {
            time: date.getTime(),
            availability: typeof availability === "number" && Number.isFinite(availability)
                ? clamp(availability, 0, 100)
                : null,
        });
    }
    const fromDate = parseDateInput(fromValue);
    const toDate = parseDateInput(toValue);
    if (!fromDate || !toDate)
        return [];
    const points = [];
    for (let date = new Date(fromDate); date.getTime() <= toDate.getTime(); date = addDays(date, 1)) {
        const key = toDateInputValue(date);
        points.push(pointsByDate.get(key) ?? {
            time: date.getTime(),
            availability: null,
        });
    }
    return points;
}
async function fetchJSON(path) {
    const response = await fetch(new URL(path, API_BASE), {
        headers: { Accept: "application/json" },
    });
    if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${await response.text()}`);
    }
    return (await response.json());
}
function renderSite(site, latestStat) {
    const name = getElement("#site-name");
    const url = getElement("#site-url");
    const status = getElement("#site-status");
    const availability = getElement("#site-availability");
    const lastCheck = getElement("#site-last-check");
    if (!name || !url || !status || !availability || !lastCheck) {
        throw new Error("Site page is missing required elements");
    }
    document.title = `${site.name} — PingTower`;
    name.textContent = site.name;
    url.href = site.url;
    url.textContent = site.url;
    if (!latestStat) {
        status.textContent = "No data";
        availability.textContent = "Availability: —";
        lastCheck.textContent = "";
        return;
    }
    status.textContent = latestStat.is_up ? "Available" : "Unavailable";
    availability.textContent =
        `Latest check: ${latestStat.is_up ? "successful" : "failed"}`;
    lastCheck.textContent =
        `Last checked: ${new Date(latestStat.checked_at).toLocaleString()}`;
}
function drawChart(points, fromDate, toDate) {
    const canvas = getElement("#site-stats-chart");
    const scroll = getElement("#chart-scroll");
    const inner = getElement(".chart-inner");
    const tooltip = getElement("#chart-tooltip");
    const message = getElement("#chart-message");
    const context = canvas?.getContext("2d");
    if (!canvas || !scroll || !inner || !tooltip || !message || !context) {
        console.error("Chart elements are missing from the page");
        return;
    }
    const visibleWidth = scroll.clientWidth;
    const startTime = startOfDay(fromDate).getTime();
    const endTime = endOfDay(toDate).getTime();
    const timeRange = Math.max(1, endTime - startTime);
    const totalDays = Math.max(1, Math.ceil(timeRange / DAY_MS));
    if (visibleWidth <= 0) {
        message.textContent = "Chart is not visible.";
        return;
    }
    const width = Math.max(visibleWidth, visibleWidth * (totalDays / 7));
    inner.style.width = `${width}px`;
    inner.style.minWidth = `${width}px`;
    canvas.style.width = `${width}px`;
    const height = 300;
    const pixelRatio = window.devicePixelRatio || 1;
    canvas.style.height = `${height}px`;
    canvas.width = Math.round(width * pixelRatio);
    canvas.height = Math.round(height * pixelRatio);
    context.setTransform(pixelRatio, 0, 0, pixelRatio, 0, 0);
    context.clearRect(0, 0, width, height);
    const padding = { left: 16, right: 56, top: 18, bottom: 42 };
    const plotLeft = padding.left;
    const plotRight = width - padding.right;
    const plotTop = padding.top;
    const plotBottom = height - padding.bottom;
    const plotWidth = plotRight - plotLeft;
    const plotHeight = plotBottom - plotTop;
    context.font = "12px sans-serif";
    context.lineWidth = 1;
    context.strokeStyle = "#a1a1a6";
    context.fillStyle = "#666";
    context.textBaseline = "middle";
    context.textAlign = "left";
    for (const percent of [0, 25, 50, 75, 100]) {
        const y = plotBottom - (percent / 100) * plotHeight;
        context.beginPath();
        context.moveTo(plotLeft, y);
        context.lineTo(plotRight, y);
        context.stroke();
        context.fillText(`${percent}%`, plotRight + 8, y);
    }
    const tickEveryDays = totalDays > 180 ? 30 :
        totalDays > 90 ? 14 :
            totalDays > 45 ? 7 :
                totalDays > 14 ? 3 : 1;
    context.textAlign = "center";
    context.textBaseline = "top";
    for (let day = 0; day <= totalDays; day += tickEveryDays) {
        const time = startTime + day * DAY_MS;
        const x = plotLeft + ((time - startTime) / timeRange) * plotWidth;
        context.beginPath();
        context.moveTo(x, plotTop);
        context.lineTo(x, plotBottom);
        context.stroke();
        context.fillText(new Date(time).toLocaleDateString("en-US", {
            month: "short",
            day: "numeric",
        }), x, plotBottom + 10);
    }
    const positions = points
        .filter((point) => point.time >= startTime &&
        point.time <= endTime)
        .map((point) => ({
        point,
        x: plotLeft + ((point.time - startTime) / timeRange) * plotWidth,
        y: point.availability == null
            ? null
            : plotBottom - (point.availability / 100) * plotHeight,
    }));
    const accent = getComputedStyle(document.documentElement)
        .getPropertyValue("--color-accent")
        .trim() || "#0071e3";
    context.strokeStyle = accent;
    context.lineWidth = 2;
    context.beginPath();
    let segmentOpen = false;
    for (const position of positions) {
        if (position.y === null) {
            continue;
        }
        if (!segmentOpen) {
            context.moveTo(position.x, position.y);
            segmentOpen = true;
        }
        else {
            context.lineTo(position.x, position.y);
        }
    }
    context.stroke();
    if (positions.every((position) => position.point.availability === null)) {
        message.textContent = "No availability data for the selected period.";
    }
    else {
        message.textContent = "";
    }
    canvas.onmousemove = (event) => {
        const rect = canvas.getBoundingClientRect();
        const mouseX = event.clientX - rect.left;
        let nearest;
        let nearestDistance = Number.POSITIVE_INFINITY;
        for (const position of positions) {
            if (position.point.availability === null)
                continue;
            const distance = Math.abs(position.x - mouseX);
            if (distance < nearestDistance) {
                nearest = position;
                nearestDistance = distance;
            }
        }
        if (!nearest || nearestDistance > 24) {
            tooltip.hidden = true;
            return;
        }
        tooltip.textContent =
            `${nearest.point.availability.toFixed(2)}% — ` +
                new Date(nearest.point.time).toLocaleString();
        tooltip.style.left = `${event.clientX + 12}px`;
        tooltip.style.top = `${event.clientY + 12}px`;
        tooltip.hidden = false;
    };
    canvas.onmouseleave = () => {
        tooltip.hidden = true;
    };
}
function setupThemeToggle() {
    const button = document.querySelector("#theme-toggle");
    if (!button) {
        console.error("Theme button not found");
        return;
    }
    document.documentElement.dataset.theme =
        localStorage.getItem("theme") === "dark" ? "dark" : "light";
    button.addEventListener("click", () => {
        const next = document.documentElement.dataset.theme === "dark" ? "light" : "dark";
        document.documentElement.dataset.theme = next;
        localStorage.setItem("theme", next);
        console.log("Site theme:", next);
    });
}
function getElement(selector) {
    return document.querySelector(selector);
}
function setText(selector, value) {
    const element = getElement(selector);
    if (element)
        element.textContent = value;
}
function parseDateInput(value) {
    const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(value);
    if (!match)
        return null;
    const year = Number(match[1]);
    const month = Number(match[2]);
    const day = Number(match[3]);
    const date = new Date(year, month - 1, day);
    if (date.getFullYear() !== year ||
        date.getMonth() !== month - 1 ||
        date.getDate() !== day) {
        return null;
    }
    return date;
}
function toDateInputValue(date) {
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, "0");
    const day = String(date.getDate()).padStart(2, "0");
    return `${year}-${month}-${day}`;
}
function startOfDay(date) {
    const result = new Date(date);
    result.setHours(0, 0, 0, 0);
    return result;
}
function endOfDay(date) {
    const result = new Date(date);
    result.setHours(23, 59, 59, 999);
    return result;
}
function addDays(date, amount) {
    const result = new Date(date);
    result.setDate(result.getDate() + amount);
    return result;
}
function clamp(value, min, max) {
    return Math.max(min, Math.min(max, value));
}
document.addEventListener("DOMContentLoaded", () => {
    setupDateControls();
    if (!siteID) {
        console.error("Missing site ID");
        setText("#site-error", "Site ID is missing.");
        return;
    }
    void loadSitePage(siteID);
});
setupThemeToggle();
