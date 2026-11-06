interface LatestStat {
  is_up: boolean;
  response_time_ms: number | null;
  checked_at: string;
}

interface Site {
  id: number;
  name: string;
  url: string;
  enabled: boolean;
  created_at: string;
}



const PAGE_SIZE = 20;

const grid = document.querySelector<HTMLDivElement>("#tile-grid");
const statusMessage = document.querySelector<HTMLParagraphElement>("#status-message");
const loadMoreButton =
  document.querySelector<HTMLButtonElement>("#load-more");

if (!grid || !statusMessage || !loadMoreButton) {
  throw new Error("Dashboard: не найдены необходимые элементы страницы");
}

let offset = 0;
let isLoading = false;
let hasMore = true;

async function loadSites(): Promise<void> {
  if (isLoading || !hasMore) return;

  isLoading = true;
  loadMoreButton.disabled = true;
  statusMessage.textContent = "Загружаем сайты…";

  const params = new URLSearchParams({
    limit: String(PAGE_SIZE),
    offset: String(offset),
  });

  try {
    const response = await fetch(`/sites?${params.toString()}`, {
      headers: { Accept: "application/json" },
    });

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`);
    }

    const sites: Site[] = await response.json();

    const sitesWithStats = await Promise.all(
      sites.map(async (site) => {
        try {
          return { site, latestStat: await getLatestStat(site.id) };
        } catch (error) {
          console.error(`Latest-stat request failed for site ${site.id}:`, error);
          return { site, latestStat: null };
        }
      }),
    );


    if (sites.length === 0 && offset === 0) {
      statusMessage.textContent = "Сайтов пока нет.";
      loadMoreButton.hidden = true;
      return;
    }

    const fragment = document.createDocumentFragment();

    for (const { site, latestStat } of sitesWithStats) {
      const tile = document.createElement("article");
      tile.className = "tile";

      const title = document.createElement("h2");
      title.className = "tile__title";
      title.textContent = site.name;

      const statusDot = document.createElement("span");
      statusDot.className = "tile__status-dot";
      statusDot.setAttribute("role", "img");

      if (!latestStat) {
        statusDot.classList.add("tile__status-dot--unknown");
        statusDot.setAttribute("aria-label", "No data");
      } else if (latestStat.is_up) {
        statusDot.classList.add("tile__status-dot--up");
        statusDot.setAttribute("aria-label", "Available");
      } else {
        statusDot.classList.add("tile__status-dot--down");
        statusDot.setAttribute("aria-label", "Unavailable");
      }

      const responseTime = document.createElement("p");
      responseTime.className = "tile__response-time";
      responseTime.textContent =
        latestStat?.response_time_ms != null
          ? `${latestStat.response_time_ms} ms`
          : "— ms";

      const lastCheck = document.createElement("p");
      lastCheck.className = "tile__last-check";
      lastCheck.textContent = latestStat
        ? `Checked at: ${new Date(latestStat.checked_at).toLocaleString()}`
        : "No data";

      tile.append(title, statusDot, responseTime, lastCheck);
      fragment.appendChild(tile);
    }


    grid.appendChild(fragment);

    offset += sites.length;
    hasMore = sites.length === PAGE_SIZE;

    statusMessage.textContent = "";
    loadMoreButton.hidden = !hasMore;
  } catch (error) {
    console.error("Не удалось загрузить сайты:", error);
    statusMessage.textContent = "Не удалось загрузить сайты.";
    loadMoreButton.hidden = false;
  } finally {
    isLoading = false;
    loadMoreButton.disabled = false;
  }
}

loadMoreButton.addEventListener("click", () => {
  void loadSites();
});

void loadSites();

const themeToggle = document.querySelector<HTMLButtonElement>("#theme-toggle");

if (themeToggle) {
  const savedTheme = localStorage.getItem("theme");
  const initialTheme = savedTheme === "dark" ? "dark" : "light";

  document.documentElement.dataset.theme = initialTheme;

  themeToggle.addEventListener("click", () => {
    const nextTheme =
      document.documentElement.dataset.theme === "dark" ? "light" : "dark";

    document.documentElement.dataset.theme = nextTheme;
    localStorage.setItem("theme", nextTheme);
  });
}

async function getLatestStat(siteID: number): Promise<LatestStat | null> {
  const response = await fetch(
    `/stats/sites/${encodeURIComponent(siteID)}/latest-stat`,
    { headers: { Accept: "application/json" } },
  );

  if (!response.ok) {
    const body = await response.text();
    throw new Error(`HTTP ${response.status}: ${body}`);
  }

  return response.json() as Promise<LatestStat | null>;
}
