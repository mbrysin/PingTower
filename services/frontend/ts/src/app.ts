interface LatestStat {
  is_up: boolean;
  response_time_ms: number | null;
}

interface Site {
  id: number;
  name: string;
  url: string;
  enabled: boolean;
  created_at: string;
  latest_stat: LatestStat | null;
}


const pageSize = 20;
let offset = 0;
let isLoading = false;
let hasMore = true;

const grid = document.querySelector<HTMLDivElement>("#tile-grid");
const message = document.querySelector<HTMLParagraphElement>("#load-message");
const loadMoreButton =
  document.querySelector<HTMLButtonElement>("#load-more");

if (!grid || !message || !loadMoreButton) {
  throw new Error("Не найдены элементы страницы для списка сайтов");
}

async function loadSites(): Promise<void> {
  if (isLoading || !hasMore) return;

  isLoading = true;
  loadMoreButton.disabled = true;
  message.textContent = "Загружаем сайты…";

  try {
    const params = new URLSearchParams({
      limit: String(pageSize),
      offset: String(offset),
    });

    const response = await fetch(`/sites?${params}`);

    if (!response.ok) {
      throw new Error(`Сервер вернул ${response.status}`);
    }

    const sites: Site[] = await response.json();
    const fragment = document.createDocumentFragment();

    for (const site of sites) {
      const tile = document.createElement("article");
      tile.className = "tile";

      const name = document.createElement("h2");
      name.textContent = site.name;

      const url = document.createElement("p");
      url.textContent = site.url;

      const status = document.createElement("p");
      status.textContent = site.enabled ? "Включён" : "Отключён";

      const statusDot = document.createElement("span");
      statusDot.className = "tile__status-dot";
      statusDot.setAttribute("role", "img");

      if (site.latest_stat === null) {
        statusDot.classList.add("tile__status-dot--unknown");
        statusDot.setAttribute("aria-label", "Нет данных");
      } else if (site.latest_stat.is_up) {
        statusDot.classList.add("tile__status-dot--up");
        statusDot.setAttribute("aria-label", "Сайт работает");
      } else {
        statusDot.classList.add("tile__status-dot--down");
        statusDot.setAttribute("aria-label", "Сайт недоступен");
      }

      tile.append(name, url, status);
      fragment.appendChild(tile);
    }

    grid.appendChild(fragment);

    offset += sites.length;
    hasMore = sites.length === pageSize;

    message.textContent = sites.length === 0
      ? (offset === 0 ? "Сайтов пока нет." : "Больше сайтов нет.")
      : "";

    loadMoreButton.hidden = !hasMore;
  } catch (error) {
    message.textContent = "Не удалось загрузить сайты. Попробуйте ещё раз.";
    console.error("Ошибка загрузки сайтов:", error);
  } finally {
    isLoading = false;
    loadMoreButton.disabled = false;
  }
}

loadMoreButton.addEventListener("click", () => {
  void loadSites();
});

void loadSites();
