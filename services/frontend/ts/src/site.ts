interface Site {
  id: number;
  name: string;
  url: string;
  enabled: boolean;
  created_at: string;
}

interface LatestStat {
  is_up: boolean;
  response_time_ms?: number;
  checked_at: string;
}

interface Stat {
  checked_at: string;
  response_time_ms?: number | null;
}

const siteID = document.body.dataset.siteId;
const STATS_API = "http://localhost:8082";

if (!siteID) {
  throw new Error("Missing site ID on page");
}

void loadSitePage(siteID);
setupThemeToggle();

async function loadSitePage(id: string): Promise<void> {
  const encodedID = encodeURIComponent(id);
  console.log("Loading site page, id:", id);

  try {
    const [site, latestStat] = await Promise.all([
      fetchJSON<Site>(`/sites/${encodedID}`),
      fetchJSON<LatestStat | null>(
        `/stats/sites/${encodedID}/latest-stat`,
      ),
    ]);

    console.log("Site response:", site);
    console.log("Latest stat response:", latestStat);

    renderSite(site, latestStat);
  } catch (error) {
    console.error("Site page loading/rendering failed:", error);
    setText(
      "#site-error",
      error instanceof Error ? error.message : String(error),
    );
    return;
  }

  try {
    const stats = await loadStats(encodedID);
    console.log("Stats response:", stats);
    drawChart(stats);
  } catch (error) {
    console.error("Chart loading failed:", error);
    setText("#chart-message", "Could not load response time history.");
  }
}


async function loadStats(siteID: string): Promise<Stat[]> {
  const to = new Date();
  const from = new Date(to.getTime() - 24 * 60 * 60 * 1000);

  const params = new URLSearchParams({
    from: from.toISOString(),
    to: to.toISOString(),
  });

  return fetchJSON<Stat[]>(
    `/sites/${siteID}/stats?${params.toString()}`,
  );
}

async function fetchJSON<T>(url: string): Promise<T> {
  const response = await fetch(url, {
    headers: { Accept: "application/json" },
  });

  if (!response.ok) {
    const body = await response.text();
    throw new Error(`HTTP ${response.status}: ${body}`);
  }

  return response.json() as Promise<T>;
}

function renderSite(site: Site, latestStat: LatestStat | null): void {
  const name = document.querySelector<HTMLElement>("#site-name");
  const url = document.querySelector<HTMLAnchorElement>("#site-url");
  const status = document.querySelector<HTMLElement>("#site-status");
  const responseTime =
    document.querySelector<HTMLElement>("#site-response-time");
  const lastCheck =
    document.querySelector<HTMLElement>("#site-last-check");

  if (!name || !url || !status || !responseTime || !lastCheck) {
    throw new Error("Site page is missing required elements");
  }

  document.title = `${site.name} — PingTower`;
  name.textContent = site.name;

  url.href = site.url;
  url.textContent = site.url;

  if (!latestStat) {
    status.textContent = "No data";
    responseTime.textContent = "—";
    lastCheck.textContent = "";
    return;
  }

  status.textContent = latestStat.is_up ? "Available" : "Unavailable";

  responseTime.textContent =
    typeof latestStat.response_time_ms === "number"
      ? `${latestStat.response_time_ms} ms`
      : "—";

  lastCheck.textContent =
    `Last checked: ${new Date(latestStat.checked_at).toLocaleString()}`;
}

function drawChart(stats: Stat[]) {
  const canvas = document.querySelector<HTMLCanvasElement>("#site-stats-chart");
  const scroll = document.querySelector<HTMLElement>("#chart-scroll");
  const inner = document.querySelector<HTMLElement>(".chart-inner");
  const yAxis = document.querySelector<HTMLElement>("#chart-y-axis");
  const tooltip = document.querySelector<HTMLElement>("#chart-tooltip");
  const message = document.querySelector<HTMLElement>("#chart-message");
  const context = canvas?.getContext("2d");

  if (!canvas || !scroll || !inner || !yAxis || !tooltip || !message || !context) {
    return;
  }

  const points = stats
    .filter(
      (stat): stat is Stat & { response_time_ms: number } =>
        typeof stat.response_time_ms === "number" &&
        Number.isFinite(stat.response_time_ms),
    )
    .map((stat) => ({
      ...stat,
      time: new Date(stat.checked_at).getTime(),
    }))
    .filter((stat) => Number.isFinite(stat.time))
    .sort((a, b) => a.time - b.time);

  if (points.length === 0) {
    message.textContent = "No response time history available.";
    tooltip.hidden = true;
    yAxis.replaceChildren();
    return;
  }

  message.textContent = "";

  const height = 280;
  const ratio = window.devicePixelRatio || 1;
  const visibleWidth = scroll.clientWidth;

  if (visibleWidth === 0) return;

  const chartWidth = Math.max(visibleWidth * 8, 1600);

  inner.style.width = `${chartWidth}px`;
  inner.style.minWidth = `${chartWidth}px`;
  canvas.style.width = `${chartWidth}px`;
  canvas.style.height = `${height}px`;
  canvas.width = Math.round(chartWidth * ratio);
  canvas.height = Math.round(height * ratio);

  context.setTransform(ratio, 0, 0, ratio, 0, 0);
  context.clearRect(0, 0, chartWidth, height);

  const padding = { top: 16, right: 56, bottom: 38, left: 12 };
  const plotLeft = padding.left;
  const plotRight = chartWidth - padding.right;
  const plotTop = padding.top;
  const plotBottom = height - padding.bottom;
  const plotWidth = plotRight - plotLeft;
  const plotHeight = plotBottom - plotTop;

  const endTime = Date.now();
  const startTime = endTime - 24 * 60 * 60 * 1000;

  const values = points.map((point) => point.response_time_ms);
  const dataMin = Math.min(...values);
  const dataMax = Math.max(...values);
  const dataRange = dataMax - dataMin;

  function getNiceStep(range: number, targetIntervals = 5): number {
    if (range <= 0) return 1;

    const roughStep = range / targetIntervals;
    const magnitude = 10 ** Math.floor(Math.log10(roughStep));
    const normalized = roughStep / magnitude;

    const niceNumber =
      normalized <= 1 ? 1 :
      normalized <= 2 ? 2 :
      normalized <= 5 ? 5 : 10;

    return niceNumber * magnitude;
  }

  const step = getNiceStep(dataRange);
  const minTick = Math.floor(dataMin / step) * step;
  const maxTick = Math.ceil(dataMax / step) * step;
  const yMin = minTick === maxTick ? minTick - step : minTick;
  const yMax = minTick === maxTick ? maxTick + step : maxTick;
  const yRange = yMax - yMin;

  context.font = "12px -apple-system, BlinkMacSystemFont, Segoe UI, sans-serif";
  context.lineWidth = 1;
  context.strokeStyle = "#a1a1a6";
  context.fillStyle = "#a1a1a6";
  context.setLineDash([]);

  yAxis.replaceChildren();
  yAxis.style.height = `${height}px`;

  for (let value = yMin; value <= yMax + step * 0.001; value += step) {
    const y = plotBottom - ((value - yMin) / yRange) * plotHeight;

    context.beginPath();
    context.moveTo(plotLeft, y);
    context.lineTo(plotRight, y);
    context.stroke();

    const label = document.createElement("span");
    label.className = "chart-y-axis__label";
    label.textContent = String(Math.round(value));
    label.style.top = `${y}px`;
    yAxis.append(label);
  }

  const halfHour = 30 * 60 * 1000;
  const firstTick = Math.ceil(startTime / halfHour) * halfHour;

  for (let time = firstTick; time <= endTime; time += halfHour) {
    const x =
      plotLeft +
      ((time - startTime) / (endTime - startTime)) * plotWidth;

    context.setLineDash([4, 4]);
    context.beginPath();
    context.moveTo(x, plotTop);
    context.lineTo(x, plotBottom);
    context.stroke();

    const date = new Date(time);

    if (date.getMinutes() === 0) {
      const hour = date.getHours();
      const label =
        hour === 0 ? "12 AM" :
        hour === 12 ? "12 PM" :
        hour < 12 ? String(hour) :
        String(hour - 12);

      context.setLineDash([]);
      context.textAlign = "center";
      context.textBaseline = "top";
      context.fillText(label, x, plotBottom + 10);
    }
  }

  context.setLineDash([]);

  const pointPositions = points.map((point) => {
    const x =
      plotLeft +
      ((point.time - startTime) / (endTime - startTime)) * plotWidth;
    const y =
      plotBottom -
      ((point.response_time_ms - yMin) / yRange) * plotHeight;

    return { point, x, y };
  });

  context.beginPath();
  context.strokeStyle =
    getComputedStyle(document.documentElement)
      .getPropertyValue("--color-accent")
      .trim() || "#0071e3";
  context.lineWidth = 2;

  pointPositions.forEach(({ x, y }, index) => {
    if (index === 0) context.moveTo(x, y);
    else context.lineTo(x, y);
  });

  context.stroke();

  scroll.scrollLeft = scroll.scrollWidth;

  canvas.onmousemove = (event: MouseEvent) => {
    const rect = canvas.getBoundingClientRect();
    const mouseX = event.clientX - rect.left;

    let nearest = pointPositions[0];
    let nearestDistance = Math.abs(nearest.x - mouseX);

    for (const position of pointPositions) {
      const distance = Math.abs(position.x - mouseX);
      if (distance < nearestDistance) {
        nearest = position;
        nearestDistance = distance;
      }
    }

    if (nearestDistance > 18) {
      tooltip.hidden = true;
      return;
    }

    tooltip.textContent = `${nearest.point.response_time_ms} ms`;
    tooltip.style.left = `${event.clientX + 12}px`;
    tooltip.style.top = `${event.clientY + 12}px`;
    tooltip.hidden = false;
  };

  canvas.onmouseleave = () => {
    tooltip.hidden = true;
  };
}

function setupThemeToggle(): void {
  const themeToggle =
    document.querySelector<HTMLButtonElement>("#theme-toggle");

  if (!themeToggle) return;

  const savedTheme = localStorage.getItem("theme");
  document.documentElement.dataset.theme =
    savedTheme === "dark" ? "dark" : "light";

  themeToggle.addEventListener("click", () => {
    const nextTheme =
      document.documentElement.dataset.theme === "dark" ? "light" : "dark";

    document.documentElement.dataset.theme = nextTheme;
    localStorage.setItem("theme", nextTheme);
  });
}
