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
  throw new Error("Dashboard: not found needed page's elements");
}

let offset = 0;
let isLoading = false;
let hasMore = true;

async function loadSites(): Promise<void> {
  if (isLoading || !hasMore) return;

  isLoading = true;
  loadMoreButton.disabled = true;
  statusMessage.textContent = "Загрузка…";

  const params = new URLSearchParams({
    limit: String(PAGE_SIZE),
    offset: String(offset),
  });

  try {
    const sitesResponse = await fetch(`/sites?${params.toString()}`, {
      headers: { Accept: "application/json" },
    });

    if (!sitesResponse.ok) {
      throw new Error(`Ошибка загрузки сайтов: HTTP ${sitesResponse.status}`);
    }

    const sites: Site[] = await sitesResponse.json();

    if (sites.length === 0 && offset === 0) {
      statusMessage.textContent = "Сайтов пока нет.";
      loadMoreButton.hidden = true;
      return;
    }

    let latestStats: Record<string, LatestStat> = {};

    if (sites.length > 0) {
      const statsParams = new URLSearchParams({
        site_ids: sites.map((site) => site.id).join(","),
      });

      const statsResponse = await fetch(
        `/stats/latest-stats?${statsParams.toString()}`,
        { headers: { Accept: "application/json" } },
      );

      if (!statsResponse.ok) {
        throw new Error(
          `Ошибка загрузки статистики: HTTP ${statsResponse.status}`,
        );
      }

      latestStats = await statsResponse.json();
    }

    const fragment = document.createDocumentFragment();

    for (const site of sites) {
      const latestStat = latestStats[String(site.id)] ?? null;

      const tile = document.createElement("article");
      tile.className = "tile";

      const title = document.createElement("a");
      title.className = "tile__title";
      title.textContent = site.name;
      title.href = `/site/${encodeURIComponent(site.id)}`;

      const statusDot = document.createElement("span");
      statusDot.className = "tile__status-dot";
      statusDot.setAttribute("role", "img");

      if (!latestStat) {
        statusDot.classList.add("tile__status-dot--unknown");
        statusDot.setAttribute("aria-label", "Нет данных");
      } else if (latestStat.is_up) {
        statusDot.classList.add("tile__status-dot--up");
        statusDot.setAttribute("aria-label", "Доступен");
      } else {
        statusDot.classList.add("tile__status-dot--down");
        statusDot.setAttribute("aria-label", "Недоступен");
      }

      const responseTime = document.createElement("p");
      responseTime.className = "tile__response-time";

      if (latestStat?.response_time_ms != null) {
        const value = document.createElement("span");
        value.textContent = String(latestStat.response_time_ms);

        const unit = document.createElement("span");
        unit.className = "tile__response-unit";
        unit.textContent = " ms";

        responseTime.append(value, unit);
      } else {
        responseTime.textContent = "—";
      }

      const lastCheck = document.createElement("p");
      lastCheck.className = "tile__last-check";
      lastCheck.textContent = latestStat
        ? `Проверено: ${new Date(latestStat.checked_at).toLocaleString()}`
        : "Нет данных";

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
    loadMoreButton.disabled = !hasMore;
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
