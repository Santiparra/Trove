import { afterEach, beforeEach, expect, it, vi } from "vitest";
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import {
  Info,
  NovelPreview,
} from "../../bindings/github.com/Santiparra/Trove/internal/desktop/models";
import {
  Novel,
  Volume,
  CatalogMatch,
} from "../../bindings/github.com/Santiparra/Trove/internal/novels/models";
import {
  Series,
  Publication,
  Candidate,
  Release,
} from "../../bindings/github.com/Santiparra/Trove/internal/novelcatalog/models";
import {
  Edition,
  Cycle,
} from "../../bindings/github.com/Santiparra/Trove/internal/books/models";
import NovelWorkspace from "./NovelWorkspace";
const covers = vi.hoisted(() => ({ CoverDataURL: vi.fn() }));

const api = vi.hoisted(() => ({
  Dashboard: vi.fn(),
  SetCompleted: vi.fn(),
  List: vi.fn(),
  Get: vi.fn(),
  Save: vi.fn(),
  ChangeStatus: vi.fn(),
  StartVolume: vi.fn(),
  FinishVolume: vi.fn(),
  SaveVolume: vi.fn(),
  Delete: vi.fn(),
  SearchCatalog: vi.fn(),
  OpenCatalog: vi.fn(),
  BrowseCatalog: vi.fn(),
  PrepareCatalog: vi.fn(),
  CatalogReleases: vi.fn(),
  PreviewPublication: vi.fn(),
  AddCatalog: vi.fn(),
  LinkCatalog: vi.fn(),
  UnlinkCatalog: vi.fn(),
  RefreshCatalog: vi.fn(),
}));
vi.mock(
  "../../bindings/github.com/Santiparra/Trove/internal/desktop/novels",
  () => api,
);
vi.mock(
  "../../bindings/github.com/Santiparra/Trove/internal/desktop/catalog",
  () => covers,
);
const info = new Info({
  statuses: [
    { value: "to-read", label: "Por leer" },
    { value: "up-next", label: "A continuación" },
    { value: "reading", label: "Leyendo" },
    { value: "read", label: "Leído" },
  ],
  formats: [{ value: "", label: "Sin especificar" }],
});
const series = new Series({
  id: "10",
  title: "Classroom of the Elite",
  authors: [{ id: "40", name: "Provider Author" }],
  language: "en",
  format: "print",
  fetchedAt: "2026-01-01T00:00:00Z",
  books: [
    {
      id: "20",
      number: 1,
      order: 1,
      title: "Volume 1",
      originalTitle: "",
      coverURL: "",
    },
    {
      id: "21",
      number: 2,
      order: 2,
      title: "Volume 2",
      originalTitle: "",
      coverURL: "",
    },
  ],
});
const publication = (number = 1) =>
  new Publication({
    seriesID: "10",
    bookID: number === 1 ? "20" : "21",
    releaseID: number === 1 ? "30" : "31",
    number,
    order: number,
    title: `Volume ${number}`,
    language: "en",
    releaseFormat: "print",
    format: "paperback",
    isbn: "9780547928227",
    description: `Enriched synopsis ${number}`,
    coverURL: "https://images.ranobedb.org/cover.jpg",
    publisher: "Synthetic publisher",
    pages: 240,
    evidence: "ranobedb-main",
    fetchedAt: "2026-01-01T00:00:00Z",
  });
const volume = (number = 1) =>
  new Volume({
    id: `volume-${number}`,
    number,
    catalog: publication(number),
    edition: new Edition({
      id: `edition-${number}`,
      language: "en",
      format: "paperback",
      identifier: "9780547928227",
    }),
    cycle: new Cycle({
      id: `cycle-${number}`,
      status: "reading",
      startedOn: "2020-01-01",
    }),
  });
const local = (values: Partial<Novel> = {}) =>
  new Novel({
    id: "local-1",
    title: "Personal title",
    authors: ["Provider Author"],
    catalog: series,
    status: "reading",
    completed: 0,
    total: null,
    notes: "Personal notes",
    volumes: [volume()],
    ...values,
  });
const preview = (values: Partial<NovelPreview> = {}) =>
  new NovelPreview({
    token: "selection-1",
    series,
    publication: publication(),
    match: new CatalogMatch({ kind: "new" }),
    ...values,
  });
const candidate = new Candidate({ id: "10", title: series.title });
function mount(localOnly = false) {
  render(
    <NovelWorkspace
      info={info}
      localOnly={localOnly}
      onBusy={vi.fn()}
      onDirty={vi.fn()}
      canLeave={() => true}
      onNotice={vi.fn()}
    />,
  );
}
async function chooseRemote() {
  fireEvent.change(screen.getByRole("searchbox", { name: "Buscar novelas" }), {
    target: { value: "Classroom of the Elite" },
  });
  fireEvent.click(await screen.findByRole("button", { name: series.title }));
  await screen.findByRole("region", { name: "Detalle de novela de catálogo" });
  await screen.findByText("Enriched synopsis 1");
}
async function openLocal(n: Novel, localOnly = false) {
  api.List.mockResolvedValue([n]);
  api.Get.mockResolvedValue(n);
  mount(localOnly);
  fireEvent.click(
    await screen.findByRole("button", { name: `Abrir novela ${n.title}` }),
  );
  await screen.findByRole("region", { name: "Detalle de novela" });
}
beforeEach(() => {
  vi.resetAllMocks();
  covers.CoverDataURL.mockResolvedValue("");
  api.List.mockResolvedValue([]);
  api.Dashboard.mockImplementation(async () => {
    const items: Novel[] = await api.List("", "");
    return {
      total: items.length,
      counts: [],
      reading: items
        .filter((n) => n.status === "reading")
        .map((n) => ({
          ...n,
          number: (n.completed || 0) + 1,
          label: String((n.completed || 0) + 1),
          coverURL: "",
        })),
      recent: [],
      upNext: [],
    };
  });
  api.SearchCatalog.mockResolvedValue({
    items: [candidate],
    total: 1,
    page: 1,
    pages: 1,
    warnings: [],
  });
  api.BrowseCatalog.mockResolvedValue(preview());
});
afterEach(cleanup);

it("busca un nombre normal y abre directamente la ficha enriquecida antes de agregar", async () => {
  const n = local();
  api.AddCatalog.mockResolvedValue(
    new CatalogMatch({ kind: "created", novel: n, message: "Agregada" }),
  );
  mount();
  await chooseRemote();
  expect(api.SearchCatalog).toHaveBeenCalledWith(
    series.title,
    "",
    1,
    "popover",
  );
  expect(api.BrowseCatalog).toHaveBeenCalledWith("10", "", "");
  expect(api.AddCatalog).not.toHaveBeenCalled();
  expect(screen.getByText("Enriched synopsis 1")).toBeTruthy();
  fireEvent.change(screen.getByLabelText("Estado al agregar novela"), {
    target: { value: "reading" },
  });
  await screen.findByRole("region", { name: "Detalle de novela" });
  expect(api.AddCatalog).toHaveBeenCalledExactlyOnceWith(
    "selection-1",
    "reading",
    0,
  );
});

it("ofrece vincular una coincidencia manual sin volver a agregarla", async () => {
  const n = local({
    catalog: null,
    title: "Manual novel",
    notes: "Keep this note",
    completed: 1,
  });
  api.BrowseCatalog.mockResolvedValue(
    preview({
      match: new CatalogMatch({
        kind: "review",
        novel: n,
        message: "ISBN local",
      }),
    }),
  );
  api.LinkCatalog.mockResolvedValue(local({ ...n, catalog: series }));
  mount();
  await chooseRemote();
  expect(screen.queryByRole("button", { name: /Agregar ·/ })).toBeNull();
  fireEvent.click(screen.getByRole("button", { name: "Vincular publicación" }));
  await screen.findByRole("region", { name: "Detalle de novela" });
  expect(api.LinkCatalog).toHaveBeenCalledExactlyOnceWith(n.id, "selection-1");
  expect(api.AddCatalog).not.toHaveBeenCalled();
  expect(screen.getByLabelText("Notas de la novela")).toHaveProperty(
    "value",
    "Keep this note",
  );
});

it("reconoce una novela ya vinculada encontrada por un alias y no ofrece agregar", async () => {
  const n = local();
  api.BrowseCatalog.mockResolvedValue(
    new NovelPreview({
      series,
      match: new CatalogMatch({ kind: "existing", novel: n }),
    }),
  );
  mount();
  fireEvent.change(screen.getByRole("searchbox"), {
    target: { value: "Classroom of the Elite" },
  });
  fireEvent.click(await screen.findByRole("button", { name: series.title }));
  await screen.findByRole("region", { name: "Detalle de novela" });
  expect(screen.queryByRole("button", { name: /Agregar ·/ })).toBeNull();
  expect(screen.getByRole("heading", { name: n.title })).toBeTruthy();
  expect(api.AddCatalog).not.toHaveBeenCalled();
});

it("cambia idioma/formato/publicación desde la ficha manteniendo el registro local", async () => {
  const n = local();
  api.PrepareCatalog.mockResolvedValue(
    preview({ match: new CatalogMatch({ kind: "existing", novel: n }) }),
  );
  const digital = new Publication({
    ...publication(),
    releaseID: "32",
    releaseFormat: "digital",
    format: "digital",
    releaseTitle: "Digital edition",
  });
  api.CatalogReleases.mockResolvedValue([
    new Release({
      id: "32",
      title: "Digital edition",
      language: "en",
      format: "digital",
      isbn: "9780547928227",
    }),
  ]);
  api.PreviewPublication.mockResolvedValue(
    preview({
      token: "selection-2",
      publication: digital,
      match: new CatalogMatch({ kind: "existing", novel: n }),
    }),
  );
  api.LinkCatalog.mockResolvedValue(n);
  await openLocal(n);
  fireEvent.click(
    screen.getByRole("button", { name: "Elegir volumen o publicación" }),
  );
  await screen.findByRole("region", { name: "Detalle de novela de catálogo" });
  fireEvent.click(
    screen.getByRole("button", { name: "Elegir volumen o publicación" }),
  );
  fireEvent.click(
    await screen.findByRole("button", { name: /Digital edition/ }),
  );
  await screen.findByText("en · Digital");
  fireEvent.click(screen.getByRole("button", { name: "Vincular publicación" }));
  await screen.findByRole("region", { name: "Detalle de novela" });
  expect(api.PreviewPublication).toHaveBeenCalledWith("selection-1", 1, "32");
  expect(api.LinkCatalog).toHaveBeenCalledWith(n.id, "selection-2");
});

it("terminar muestra metadata del siguiente sin empezarlo", async () => {
  const n = local();
  api.FinishVolume.mockResolvedValue(
    local({
      status: "up-next",
      completed: 1,
      volumes: [
        new Volume({
          ...volume(),
          cycle: new Cycle({ id: "cycle-1", status: "read" }),
        }),
      ],
    }),
  );
  api.RefreshCatalog.mockResolvedValue(
    local({
      status: "up-next",
      completed: 1,
      volumes: [
        volume(),
        new Volume({
          ...volume(2),
          cycle: new Cycle({ id: "cycle-2", status: "to-read" }),
        }),
      ],
    }),
  );
  await openLocal(n);
  fireEvent.click(screen.getByRole("button", { name: "Volumen terminado" }));
  await screen.findByText("Enriched synopsis 2");
  expect(api.RefreshCatalog).toHaveBeenCalledWith(n.id, 2, false);
  expect(api.StartVolume).not.toHaveBeenCalled();
  expect(screen.getByLabelText(`Estado de ${n.title}`)).toHaveProperty(
    "value",
    "up-next",
  );
});

it("navega por autor con su identidad de RanobeDB y resultados locales/remotos", async () => {
  const n = local();
  api.SearchCatalog.mockResolvedValue({
    items: [new Candidate({ id: "11", title: "Another novel by this author" })],
    total: 1,
    page: 1,
    pages: 1,
    warnings: [],
  });
  await openLocal(n);
  fireEvent.click(screen.getByRole("button", { name: "Provider Author" }));
  await screen.findByRole("button", { name: "Another novel by this author" });
  expect(api.List).toHaveBeenCalledWith("Provider Author", "");
  expect(api.SearchCatalog).toHaveBeenCalledWith(
    "Provider Author",
    "40",
    1,
    "list",
  );
});

it("en modo local conserva metadata guardada y no consulta fuentes externas", async () => {
  const n = local();
  await openLocal(n, true);
  expect(screen.getByText("Enriched synopsis 1")).toBeTruthy();
  fireEvent.click(screen.getByText("Identificadores y fuentes"));
  expect(
    screen.getByRole("button", { name: "Refrescar metadata" }),
  ).toHaveProperty("disabled", true);
  expect(
    screen.getByRole("button", { name: "Elegir volumen o publicación" }),
  ).toHaveProperty("disabled", true);
  expect(api.SearchCatalog).not.toHaveBeenCalled();
  expect(api.RefreshCatalog).not.toHaveBeenCalled();
});
it("deja metadata fuera de gestión rápida y las fuentes antes de la galería", async () => {
  const n = local();
  api.RefreshCatalog.mockResolvedValue(n);
  await openLocal(n);
  const sources = document.querySelector<HTMLDetailsElement>(
    ".profile-identifiers",
  )!;
  const gallery = screen.getByRole("region", {
    name: "Volúmenes de la novela",
  });
  expect(
    sources.compareDocumentPosition(gallery) & Node.DOCUMENT_POSITION_FOLLOWING,
  ).toBeTruthy();
  expect(sources.open).toBe(false);
  const metadata = screen.getByText("Catálogo y metadata");
  expect(sources.contains(metadata)).toBe(false);
  expect(
    screen
      .getByRole("complementary", { name: "Gestión del libro" })
      .contains(metadata),
  ).toBe(false);
  const refresh = screen.getByRole("button", { name: "Refrescar metadata" });
  expect(sources.open).toBe(false);
  fireEvent.click(refresh);
  await waitFor(() =>
    expect(api.RefreshCatalog).toHaveBeenCalledWith(n.id, 1, true),
  );
});
it("mantiene la galería antes de recibir metadata y al cambiar a volúmenes futuros", async () => {
  let resolve!: (value: NovelPreview) => void;
  api.BrowseCatalog.mockResolvedValue(
    new NovelPreview({
      token: "series-header",
      series,
      publication: new Publication(),
      match: new CatalogMatch({ kind: "new" }),
    }),
  );
  api.PreviewPublication.mockImplementationOnce(
    () =>
      new Promise<NovelPreview>((done) => {
        resolve = done;
      }),
  ).mockResolvedValue(
    preview({ token: "future-preview", publication: publication(2) }),
  );
  mount();
  fireEvent.change(screen.getByRole("searchbox"), {
    target: { value: series.title },
  });
  fireEvent.click(await screen.findByRole("button", { name: series.title }));
  expect(
    await screen.findByRole("button", { name: "Abrir volumen 2" }),
  ).toBeTruthy();
  expect(
    screen.getByRole("status", { name: "Cargando información del volumen" }),
  ).toBeTruthy();
  expect(
    screen.getByRole("button", { name: "Agregar · Por leer" }),
  ).toHaveProperty("disabled", true);
  resolve(preview());
  await screen.findByText("Enriched synopsis 1");
  fireEvent.click(screen.getByRole("button", { name: "Abrir volumen 2" }));
  await screen.findByText("Enriched synopsis 2");
  expect(screen.getByRole("button", { name: "Abrir volumen 1" })).toBeTruthy();
  expect(screen.getByRole("button", { name: "Abrir volumen 2" })).toBeTruthy();
  fireEvent.click(screen.getByRole("button", { name: "Abrir volumen 1" }));
  await screen.findByText("Enriched synopsis 1");
  expect(api.PreviewPublication).toHaveBeenCalledTimes(2);
  expect(api.AddCatalog).not.toHaveBeenCalled();
});
it("muestra futuros del catálogo sin registrarlos ni cambiar progreso al abrirlos", async () => {
  const n = local();
  api.PrepareCatalog.mockResolvedValue(
    preview({
      publication: publication(2),
      match: new CatalogMatch({ kind: "existing", novel: n }),
    }),
  );
  await openLocal(n);
  expect(screen.getByRole("button", { name: "Abrir volumen 2" })).toBeTruthy();
  fireEvent.click(screen.getByRole("button", { name: "Abrir volumen 2" }));
  await screen.findByRole("region", { name: "Detalle de novela de catálogo" });
  expect(api.PrepareCatalog).toHaveBeenCalledExactlyOnceWith(n.id, 2);
  expect(api.SaveVolume).not.toHaveBeenCalled();
  expect(api.FinishVolume).not.toHaveBeenCalled();
  expect(api.LinkCatalog).not.toHaveBeenCalled();
});
