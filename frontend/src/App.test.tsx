import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  cleanup,
  act,
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from "@testing-library/react";
import App from "./App";
const novelAPI = vi.hoisted(() => ({
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
  "../bindings/github.com/Santiparra/Trove/internal/desktop/novels",
  () => novelAPI,
);
import * as BackupService from "../bindings/github.com/Santiparra/Trove/internal/desktop/backups";
const catalog = vi.hoisted(() => ({
  State: vi.fn(),
  ProviderHealth: vi.fn(),
  CheckProviders: vi.fn(),
  SetLocalOnly: vi.fn(),
  SaveKey: vi.fn(),
  ClearKey: vi.fn(),
  Search: vi.fn(),
  Editions: vi.fn(),
  Preview: vi.fn(),
  OpenWork: vi.fn(),
  Add: vi.fn(),
  Link: vi.fn(),
  Refresh: vi.fn(),
  Unlink: vi.fn(),
  CoverDataURL: vi.fn(),
}));
vi.mock(
  "../bindings/github.com/Santiparra/Trove/internal/desktop/catalog",
  () => catalog,
);

vi.mock(
  "../bindings/github.com/Santiparra/Trove/internal/desktop/backups",
  () => ({
    State: vi.fn().mockResolvedValue({
      profile: "test",
      folder: "",
      phase: "unconfigured",
      pending: false,
      lastSavedAt: "",
      generation: "",
      lastError: "",
    }),
    ChooseFolder: vi.fn(),
    BackupNow: vi.fn(),
    ChooseRestore: vi.fn(),
    RestoreChoices: vi.fn(),
    PreviewGeneration: vi.fn(),
    CancelRestore: vi.fn(),
    ConfirmRestore: vi.fn(),
  }),
);

const api = vi.hoisted(() => ({
  Info: vi.fn(),
  Dashboard: vi.fn(),
  ListBooks: vi.fn(),
  ListSeries: vi.fn(),
  SaveBook: vi.fn(),
  GetBook: vi.fn(),
  SetBookClassification: vi.fn(),
  ChangeBookStatus: vi.fn(),
  DeleteBook: vi.fn(),
  SaveBookNotes: vi.fn(),
  SetBookFavorite: vi.fn(),
  SetBookOwned: vi.fn(),
  ListSeriesBooks: vi.fn(),
  SearchBooks: vi.fn(),
}));
vi.mock(
  "../bindings/github.com/Santiparra/Trove/internal/desktop/library",
  () => api,
);

const sample = {
  id: "synthetic-book",
  medium: "books",
  favorite: false,
  title: "Obra sintética",
  authors: ["Autor A"],
  series: null,
  edition: {
    id: "edition",
    language: "",
    format: "",
    identifier: "",
    owned: false,
  },
  cycle: { id: "cycle", status: "reading", startedOn: "", finishedOn: "" },
  notes: "Nota sintética",
  addedOn: "2026-01-01",
};
const statuses = [
  { value: "up-next", label: "A continuación" },
  { value: "to-read", label: "Por leer" },
  { value: "reading", label: "Leyendo" },
  { value: "read", label: "Leído" },
  { value: "paused", label: "Pausado" },
  { value: "did-not-finish", label: "Sin terminar" },
];
function overview(withBook = false) {
  return {
    medium: "books",
    total: withBook ? 1 : 0,
    counts: statuses.map((option) => ({
      status: option.value,
      count: withBook && option.value === "reading" ? 1 : 0,
    })),
    currentlyReading: withBook ? [sample] : [],
    recentlyFinished: [],
    readWithoutDate: 0,
  };
}

beforeEach(() => {
  vi.resetAllMocks();
  novelAPI.List.mockResolvedValue([]);
  novelAPI.Dashboard.mockImplementation(async () => {
    const items = await novelAPI.List("", "");
    return {
      total: items.length,
      counts: [],
      reading: items
        .filter((n: { status: string }) => n.status === "reading")
        .map(
          (n: {
            id: string;
            title: string;
            authors: string[];
            status: string;
            completed: number | null;
            total: number | null;
          }) => ({
            ...n,
            number: (n.completed || 0) + 1,
            label: String((n.completed || 0) + 1),
            coverURL: "",
          }),
        ),
      recent: [],
      upNext: [],
    };
  });
  novelAPI.SearchCatalog.mockResolvedValue({
    items: [],
    total: 0,
    page: 1,
    pages: 1,
    warnings: [],
  });
  vi.mocked(BackupService.RestoreChoices).mockResolvedValue([]);
  catalog.State.mockResolvedValue({ configured: true });
  catalog.ProviderHealth.mockResolvedValue({
    providers: [],
    checking: false,
    nextCheckAt: "",
  });
  catalog.CoverDataURL.mockResolvedValue("");
  catalog.Search.mockResolvedValue({ works: [], warnings: [] });
  vi.mocked(BackupService.State).mockResolvedValue({
    profile: "test",
    folder: "",
    phase: "unconfigured",
    pending: false,
    lastSavedAt: "",
    generation: "",
    lastError: "",
  });
  api.Info.mockResolvedValue({
    profile: "test",
    dataDir: "synthetic/test",
    buildID: "synthetic",
    statuses,
    formats: [{ value: "", label: "Sin especificar" }],
  });
  api.Dashboard.mockResolvedValue(overview());
  api.ListBooks.mockResolvedValue([]);
  api.ListSeries.mockResolvedValue([]);
  api.ListSeriesBooks.mockResolvedValue([]);
  api.SearchBooks.mockResolvedValue({ matches: [], total: 0 });
  api.GetBook.mockResolvedValue(sample);
});
afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
});

describe("Catálogo de Libros", () => {
  const snapshot = {
    provider: "hardcover",
    fetchedAt: "2026-01-01T00:00:00Z",
    work: {
      id: "10",
      title: "Obra externa",
      authors: ["Autora externa"],
      series: [],
    },
    edition: {
      id: "20",
      workId: "10",
      title: "Edición externa",
      language: "en",
      format: "digital",
      publisher: "Editorial sintética",
      pages: 240,
      publishedOn: "2020",
      declared: [],
    },
    identifiers: [],
    coverURL: "",
    description: "Descripción guardada",
    sources: {},
    warnings: [],
  };
  const linked = {
    ...sample,
    edition: { ...sample.edition, catalog: snapshot },
  };
  async function enterCatalog() {
    await screen.findByRole("button", { name: "Ver todos los libros" });
  }
  async function chooseEdition() {
    fireEvent.change(
      await screen.findByRole("searchbox", { name: "Buscar por título" }),
      { target: { value: "9780547928227" } },
    );
    fireEvent.click(
      await screen.findByRole("button", { name: /Obra externa/ }),
    );
    await screen.findByRole("region", { name: "Detalle del libro" });
  }
  beforeEach(() => {
    catalog.Search.mockResolvedValue({
      works: [{ ...snapshot.work, matchedEditionId: "20" }],
      warnings: [],
    });
    catalog.Preview.mockResolvedValue({ token: "synthetic-preview", snapshot });
  });
  it("consulta y revisa antes de agregar una edición y abre su snapshot local", async () => {
    catalog.Add.mockResolvedValue({
      kind: "added",
      book: linked,
      message: "Edición agregada.",
    });
    render(<App />);
    await enterCatalog();
    await chooseEdition();
    expect(catalog.Add).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole("button", { name: "Agregar · Por leer" }));
    await screen.findByRole("region", { name: "Detalle del libro" });
    await waitFor(() =>
      expect(catalog.Add).toHaveBeenCalledWith(
        "synthetic-preview",
        "",
        "to-read",
      ),
    );
    expect(screen.getByText("Descripción guardada")).toBeTruthy();
    expect(screen.getByRole("heading", { name: sample.title })).toBeTruthy();
    const panels = screen
      .getByLabelText("Notas personales del libro")
      .closest(".reading-panels-two");
    expect(panels).not.toBeNull();
    expect(
      screen
        .getByRole("region", { name: "Historial de lecturas" })
        .closest(".reading-panels-two"),
    ).toBe(panels);
  });
  it("pide revisar una coincidencia local antes de vincular, sin guardar un duplicado manual", async () => {
    catalog.Preview.mockResolvedValue({
      token: "synthetic-preview",
      snapshot,
      match: {
        kind: "review",
        book: sample,
        message: "Revisá la coincidencia local.",
      },
    });
    catalog.Link.mockResolvedValue(linked);
    render(<App />);
    await enterCatalog();
    await chooseEdition();
    fireEvent.click(
      await screen.findByRole("button", {
        name: "Vincular con registro local",
      }),
    );
    await screen.findByRole("region", { name: "Detalle del libro" });
    expect(catalog.Link).toHaveBeenCalledWith(sample.id, "synthetic-preview");
    expect(api.SaveBook).not.toHaveBeenCalled();
    expect(catalog.Add).not.toHaveBeenCalled();
    expect(screen.getByLabelText("Notas personales del libro")).toHaveProperty(
      "value",
      sample.notes,
    );
  });
  it("conserva snapshot y borrador de notas ante un fallo de refresco y tras refrescar", async () => {
    api.Dashboard.mockResolvedValue(overview(true));
    api.GetBook.mockResolvedValue(linked);
    catalog.Refresh.mockRejectedValueOnce(
      new Error("Catálogo no disponible"),
    ).mockResolvedValueOnce(linked);
    render(<App />);
    fireEvent.click(
      await screen.findByRole("button", { name: `Abrir ${sample.title}` }),
    );
    const notes = await screen.findByLabelText("Notas personales del libro");
    fireEvent.change(notes, { target: { value: "Borrador sin guardar" } });
    fireEvent.click(screen.getByRole("button", { name: "Refrescar metadata" }));
    await screen.findByText("Catálogo no disponible");
    expect(screen.getByText("Descripción guardada")).toBeTruthy();
    expect(notes).toHaveProperty("value", "Borrador sin guardar");
    fireEvent.click(screen.getByRole("button", { name: "Refrescar metadata" }));
    await screen.findByText("Metadata de catálogo actualizada.");
    expect(notes).toHaveProperty("value", "Borrador sin guardar");
  });
  it("en solo local muestra snapshots y permite gestión personal sin ofrecer refresco remoto", async () => {
    catalog.ProviderHealth.mockResolvedValue({
      providers: [],
      localOnly: true,
      checking: false,
      nextCheckAt: "",
    });
    api.Dashboard.mockResolvedValue(overview(true));
    api.GetBook.mockResolvedValue(linked);
    api.SetBookFavorite.mockResolvedValue({ ...linked, favorite: true });
    render(<App />);
    fireEvent.click(
      await screen.findByRole("button", { name: `Abrir ${sample.title}` }),
    );
    expect(await screen.findByText("Descripción guardada")).toBeTruthy();
    expect(
      screen.getByRole("button", { name: "Refrescar metadata" }),
    ).toHaveProperty("disabled", true);
    expect(
      screen.getByRole("button", { name: "Cambiar edición vinculada" }),
    ).toHaveProperty("disabled", true);
    expect(screen.getByLabelText("Notas personales del libro")).toHaveProperty(
      "disabled",
      false,
    );
    fireEvent.click(screen.getByRole("button", { name: "Favorito" }));
    await waitFor(() =>
      expect(api.SetBookFavorite).toHaveBeenCalledWith(
        sample.id,
        true,
        "books",
      ),
    );
    expect(catalog.Search).not.toHaveBeenCalled();
    expect(catalog.Refresh).not.toHaveBeenCalled();
  });
  it("abre directamente una ficha desde una obra, sin interponer el selector de ediciones", async () => {
    catalog.Search.mockResolvedValue({ works: [snapshot.work], warnings: [] });
    catalog.OpenWork.mockResolvedValue({
      token: "synthetic-preview",
      snapshot,
    });
    render(<App />);
    await enterCatalog();
    await chooseEdition();
    expect(catalog.OpenWork).toHaveBeenCalledWith("10");
    expect(catalog.Editions).not.toHaveBeenCalled();
    expect(screen.queryByRole("region", { name: "Ediciones" })).toBeNull();
    expect(
      screen.queryByText("HARDCOVER · EDICIONES · DATOS LOCALES"),
    ).toBeNull();
    expect(
      screen.getByRole("heading", { name: "Edición externa" }),
    ).toBeTruthy();
  });
  it("reconoce una edición local antes de ofrecer agregar y muestra su gestión", async () => {
    catalog.Preview.mockResolvedValue({
      token: "synthetic-preview",
      snapshot,
      match: { kind: "existing", book: linked, message: "Ya existe" },
    });
    render(<App />);
    await enterCatalog();
    await chooseEdition();
    expect(
      await screen.findByLabelText("Notas personales del libro"),
    ).toHaveProperty("value", sample.notes);
    expect(screen.queryByRole("button", { name: /Agregar ·/ })).toBeNull();
    expect(screen.getByLabelText(`Estado de ${sample.title}`)).toHaveProperty(
      "value",
      "reading",
    );
    expect(catalog.Add).not.toHaveBeenCalled();
  });
  it("mantiene visible un ISBN de otra edición aunque la obra esté en la biblioteca", async () => {
    api.SearchBooks.mockResolvedValue({ matches: [linked], total: 1 });
    catalog.Preview.mockResolvedValue({
      token: "synthetic-preview",
      snapshot: { ...snapshot, edition: { ...snapshot.edition, id: "21" } },
      match: { kind: "review", book: linked, message: "Otra edición" },
    });
    catalog.Search.mockResolvedValue({
      works: [{ ...snapshot.work, matchedEditionId: "21" }],
      warnings: [],
    });
    render(<App />);
    await enterCatalog();
    await chooseEdition();
    expect(catalog.Preview).toHaveBeenCalledWith("21");
    expect(
      screen.getByRole("button", { name: "Vincular con registro local" }),
    ).toBeTruthy();
    expect(catalog.Add).not.toHaveBeenCalled();
  });
  it("preselecciona una serie inequívoca y agrega directamente con el estado elegido", async () => {
    const withSeries = {
      ...snapshot,
      work: {
        ...snapshot.work,
        series: [{ id: "40", name: "Serie externa", position: 1 }],
      },
    };
    catalog.Preview.mockResolvedValue({
      token: "synthetic-preview",
      snapshot: withSeries,
    });
    catalog.Add.mockResolvedValue({
      kind: "created",
      book: linked,
      message: "Agregado",
    });
    render(<App />);
    await enterCatalog();
    await chooseEdition();
    expect(screen.getByLabelText("Serie al agregar")).toHaveProperty(
      "value",
      "40",
    );
    fireEvent.change(screen.getByLabelText("Estado al agregar"), {
      target: { value: "read" },
    });
    await screen.findByLabelText("Notas personales del libro");
    expect(catalog.Add).toHaveBeenCalledWith("synthetic-preview", "40", "read");
    expect(catalog.Add).toHaveBeenCalledOnce();
  });
  it("mantiene los resultados locales utilizables mientras el proveedor demora", async () => {
    api.SearchBooks.mockResolvedValue({ matches: [sample], total: 1 });
    catalog.Search.mockImplementation(() => new Promise(() => {}));
    render(<App />);
    await enterCatalog();
    fireEvent.change(screen.getByRole("searchbox"), {
      target: { value: "sintética" },
    });
    const local = await screen.findByRole("button", {
      name: `Abrir resultado ${sample.title}`,
    });
    await waitFor(() => expect(catalog.Search).toHaveBeenCalled());
    fireEvent.click(local);
    expect(
      await screen.findByLabelText("Notas personales del libro"),
    ).toHaveProperty("value", sample.notes);
  });
  it("busca también en catálogo al abrir una serie de una ficha remota", async () => {
    catalog.Preview.mockResolvedValue({
      token: "synthetic-preview",
      snapshot: {
        ...snapshot,
        work: {
          ...snapshot.work,
          series: [{ id: "40", name: "Serie externa", position: 1 }],
        },
      },
    });
    api.ListBooks.mockResolvedValue([sample]);
    render(<App />);
    await enterCatalog();
    await chooseEdition();
    catalog.Search.mockResolvedValue({
      works: [{ ...snapshot.work, id: "11", title: "Otro libro de la serie" }],
      warnings: [],
    });
    fireEvent.click(screen.getByRole("button", { name: "Serie externa · #1" }));
    expect(
      await screen.findByRole("button", { name: /Otro libro de la serie/ }),
    ).toBeTruthy();
    expect(api.ListBooks).toHaveBeenCalledWith("Serie externa", "", "books");
    expect(catalog.Search).toHaveBeenLastCalledWith("Serie externa");
    expect(
      screen.getByRole("button", { name: `Abrir ${sample.title}` }),
    ).toBeTruthy();
  });
  it("abre la búsqueda compartida al pulsar un autor de la ficha remota", async () => {
    render(<App />);
    await enterCatalog();
    await chooseEdition();
    catalog.Search.mockResolvedValue({
      works: [{ ...snapshot.work, id: "11", title: "Otro libro de la autora" }],
      warnings: [],
    });
    fireEvent.click(screen.getByRole("button", { name: "Autora externa" }));
    expect(
      await screen.findByRole("button", { name: /Otro libro de la autora/ }),
    ).toBeTruthy();
    expect(api.ListBooks).toHaveBeenCalledWith("Autora externa", "", "books");
    expect(catalog.Search).toHaveBeenLastCalledWith("Autora externa");
  });
  it("conserva el borrador al volver a elegir la edición local actual", async () => {
    api.Dashboard.mockResolvedValue(overview(true));
    api.GetBook.mockResolvedValue(linked);
    catalog.Editions.mockResolvedValue([snapshot.edition]);
    catalog.Preview.mockResolvedValue({
      token: "synthetic-preview",
      snapshot,
      match: { kind: "existing", book: linked, message: "Ya existe" },
    });
    const confirm = vi.spyOn(window, "confirm");
    render(<App />);
    fireEvent.click(
      await screen.findByRole("button", { name: `Abrir ${sample.title}` }),
    );
    const notes = await screen.findByLabelText("Notas personales del libro");
    fireEvent.change(notes, { target: { value: "Borrador sin guardar" } });
    fireEvent.click(
      screen.getByRole("button", { name: "Cambiar edición vinculada" }),
    );
    fireEvent.click(
      await screen.findByRole("button", { name: /Edición externa/ }),
    );
    await waitFor(() =>
      expect(screen.queryByRole("region", { name: "Ediciones" })).toBeNull(),
    );
    expect(notes).toHaveProperty("value", "Borrador sin guardar");
    expect(confirm).not.toHaveBeenCalled();
  });
  it("gestiona claves desde el encabezado sin ocupar la ficha", async () => {
    catalog.State.mockResolvedValue({ configured: false });
    catalog.SaveKey.mockResolvedValue({ configured: true });
    render(<App />);
    await enterCatalog();
    expect(screen.queryByLabelText("Clave de catálogo")).toBeNull();
    fireEvent.click(screen.getByRole("button", { name: /Claves ·/ }));
    fireEvent.change(screen.getByLabelText("Clave de catálogo"), {
      target: { value: "synthetic-key" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Guardar clave" }));
    await screen.findByText("Hardcover · Configurado");
    expect(screen.getByLabelText("Clave de catálogo")).toHaveProperty(
      "value",
      "",
    );
    expect(catalog.SaveKey).toHaveBeenCalledWith("synthetic-key");
  });
  it("permite seguir con alta manual sin clave y no ofrece catálogo en Novelas ligeras", async () => {
    catalog.State.mockResolvedValue({ configured: false });
    render(<App />);
    await enterCatalog();
    fireEvent.click(screen.getByRole("button", { name: "Añadir libro" }));
    await screen.findByRole("button", { name: "Guardar libro" });
    fireEvent.click(screen.getByRole("button", { name: "Novelas ligeras" }));
    await screen.findByRole("button", { name: "Añadir novela" });
    expect(
      screen.queryByRole("button", { name: "Buscar en catálogo" }),
    ).toBeNull();
    expect(catalog.Search).not.toHaveBeenCalled();
  });
});

describe("Novelas ligeras locales", () => {
  const novel = {
    id: "synthetic-novel",
    title: "Novela sintética",
    authors: ["Autora"],
    status: "reading",
    completed: 0,
    total: null,
    notes: "",
    favorite: false,
    addedOn: "2026-01-01",
    volumes: [],
  };
  async function openNovels() {
    novelAPI.List.mockResolvedValue([novel]);
    novelAPI.Get.mockResolvedValue(novel);
    render(<App />);
    await screen.findByRole("heading", { name: "Tu biblioteca" });
    fireEvent.click(screen.getByRole("button", { name: "Novelas ligeras" }));
    await screen.findByRole("button", { name: "Ver todas las novelas" });
  }
  it("crea una novela principal sin usar el alta de libros o volúmenes anteriores", async () => {
    const saved = {
      ...novel,
      id: "saved-novel",
      title: "Novela nueva",
      status: "to-read",
    };
    novelAPI.Save.mockResolvedValue(saved);
    await openNovels();
    expect(
      screen.queryByRole("button", { name: "Abrir Obra sintética" }),
    ).toBeNull();
    fireEvent.click(screen.getByRole("button", { name: "Añadir novela" }));
    fireEvent.change(screen.getByLabelText(/Título/), {
      target: { value: "Novela nueva" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Guardar novela" }));
    await waitFor(() => expect(novelAPI.Save).toHaveBeenCalledOnce());
    expect(novelAPI.Save.mock.calls[0][0]).toMatchObject({
      title: "Novela nueva",
      completed: 0,
    });
    expect(
      await screen.findByRole("region", { name: "Detalle de novela" }),
    ).toBeTruthy();
    expect(
      screen.getByRole("button", { name: "Editar novela y progreso" }),
    ).toBeTruthy();
    expect(api.SaveBook).not.toHaveBeenCalled();
  });
  it("busca, lista y cambia estados con el contexto explícito de novelas ligeras", async () => {
    const updated = { ...novel, status: "up-next" };
    novelAPI.ChangeStatus.mockResolvedValue(updated);
    await openNovels();
    fireEvent.change(screen.getByRole("searchbox"), {
      target: { value: "Novela" },
    });
    await waitFor(() =>
      expect(novelAPI.List).toHaveBeenCalledWith("Novela", ""),
    );
    fireEvent.click(
      await screen.findByRole("button", {
        name: "Abrir novela Novela sintética",
      }),
    );
    const select = await screen.findByLabelText("Estado de Novela sintética");
    novelAPI.List.mockResolvedValue([updated]);
    fireEvent.change(select, { target: { value: "up-next" } });
    await waitFor(() =>
      expect(novelAPI.ChangeStatus).toHaveBeenCalledWith(novel.id, "up-next"),
    );
    await waitFor(() => expect(select).toHaveProperty("value", "up-next"));
    expect(api.ChangeBookStatus).not.toHaveBeenCalled();
  });
  it("ignora una respuesta tardía de Libros después de cambiar de sección", async () => {
    let finishBooks!: (value: ReturnType<typeof overview>) => void;
    const delayed = new Promise<ReturnType<typeof overview>>((resolve) => {
      finishBooks = resolve;
    });
    api.Dashboard.mockImplementation((medium: string) =>
      medium === "books" ? delayed : Promise.resolve(overview(false)),
    );
    novelAPI.List.mockResolvedValue([novel]);
    render(<App />);
    fireEvent.click(screen.getByRole("button", { name: "Novelas ligeras" }));
    await screen.findByRole("button", {
      name: "Abrir novela Novela sintética",
    });
    await act(async () => {
      finishBooks(overview(true));
    });
    expect(
      screen.queryByRole("button", { name: "Abrir Obra sintética" }),
    ).toBeNull();
    expect(
      screen.getByRole("button", { name: "Abrir novela Novela sintética" }),
    ).toBeTruthy();
  });
});

describe("portal y navegación de libros", () => {
  it("abre el portal, no el listado, y el estado elegido define la consulta al backend", async () => {
    api.Dashboard.mockResolvedValue(overview(true));
    api.ListBooks.mockResolvedValue([sample]);
    render(<App />);
    await screen.findByRole("heading", { name: "Tu biblioteca" });
    expect(api.ListBooks).not.toHaveBeenCalled();
    expect(
      screen.queryByRole("region", { name: "Biblioteca de libros" }),
    ).toBeNull();
    fireEvent.click(screen.getByRole("button", { name: "Leyendo: 1" }));
    await waitFor(() =>
      expect(api.ListBooks).toHaveBeenCalledWith("", "reading", "books"),
    );
    fireEvent.click(
      await screen.findByRole("button", { name: "Abrir Obra sintética" }),
    );
    expect(
      await screen.findByRole("region", { name: "Detalle del libro" }),
    ).toBeTruthy();
    fireEvent.click(screen.getByRole("button", { name: "Editar libro" }));
    expect(screen.getByRole("region", { name: "Editar libro" })).toBeTruthy();
    expect(
      screen.queryByRole("region", { name: "Detalle del libro" }),
    ).toBeNull();
  });

  it("mantiene el estado confirmado si el cambio rápido falla y no utiliza guardado completo", async () => {
    api.Dashboard.mockResolvedValue(overview(true));
    api.ChangeBookStatus.mockRejectedValue(
      new Error("No se pudo cambiar el estado."),
    );
    render(<App />);
    const select = await screen.findByLabelText("Estado de Obra sintética");
    fireEvent.change(select, { target: { value: "read" } });
    expect(await screen.findByRole("alert")).toHaveProperty(
      "textContent",
      "No se pudo cambiar el estado.",
    );
    expect(select).toHaveProperty("value", "reading");
    expect(api.ChangeBookStatus).toHaveBeenCalledWith(
      sample.id,
      "read",
      "books",
    );
    expect(api.SaveBook).not.toHaveBeenCalled();
  });

  it("refresca las cantidades y lecturas del portal después de cambiar un estado", async () => {
    const updated = overview();
    updated.total = 1;
    updated.readWithoutDate = 1;
    updated.counts = updated.counts.map((item) => ({
      ...item,
      count: item.status === "read" ? 1 : 0,
    }));
    api.Dashboard.mockResolvedValueOnce(overview(true)).mockResolvedValue(
      updated,
    );
    api.ChangeBookStatus.mockResolvedValue({
      ...sample,
      cycle: { ...sample.cycle, status: "read" },
    });
    render(<App />);
    fireEvent.change(await screen.findByLabelText("Estado de Obra sintética"), {
      target: { value: "read" },
    });
    expect(
      await screen.findByRole("button", { name: "Leído: 1" }),
    ).toBeTruthy();
    expect(screen.queryByLabelText("Estado de Obra sintética")).toBeNull();
    expect(screen.getByRole("button", { name: "Leyendo: 0" })).toBeTruthy();
  });

  it("la búsqueda conserva el portal hasta elegir todos los resultados", async () => {
    render(<App />);
    await screen.findByRole("heading", { name: "Tu biblioteca" });
    fireEvent.change(screen.getByRole("searchbox"), {
      target: { value: "sintética" },
    });
    await waitFor(() =>
      expect(api.SearchBooks).toHaveBeenCalledWith("sintética", "books"),
    );
    expect(api.ListBooks).not.toHaveBeenCalled();
    expect(screen.getByRole("heading", { name: "Tu biblioteca" })).toBeTruthy();
    fireEvent.click(
      await screen.findByRole("button", {
        name: "Ver todos los resultados (0)",
      }),
    );
    await waitFor(() =>
      expect(api.ListBooks).toHaveBeenCalledWith("sintética", "", "books"),
    );
    expect(
      screen.getByRole("region", { name: "Biblioteca de libros" }),
    ).toBeTruthy();
  });
});

describe("navegación y datos personales", () => {
  it("cambia el listado mediante botones de estado, incluido A continuación", async () => {
    render(<App />);
    fireEvent.click(await screen.findByRole("button", { name: "Leyendo: 0" }));
    const tabs = await screen.findByRole("navigation", {
      name: "Estados de lectura",
    });
    fireEvent.click(
      within(tabs).getByRole("button", { name: "A continuación" }),
    );
    await waitFor(() =>
      expect(api.ListBooks).toHaveBeenCalledWith("", "up-next", "books"),
    );
    expect(screen.queryByLabelText("Filtrar por estado")).toBeNull();
  });
  it("cierra sugerencias con Escape y abre una coincidencia sin pasar por el listado", async () => {
    api.SearchBooks.mockResolvedValue({ matches: [sample], total: 1 });
    render(<App />);
    await screen.findByRole("heading", { name: "Tu biblioteca" });
    const input = screen.getByRole("searchbox");
    fireEvent.change(input, { target: { value: "Obra" } });
    const result = await screen.findByRole("button", {
      name: "Abrir resultado Obra sintética",
    });
    fireEvent.keyDown(result, { key: "Escape" });
    expect(
      screen.queryByRole("region", { name: "Resultados de búsqueda" }),
    ).toBeNull();
    fireEvent.keyDown(input, { key: "ArrowDown" });
    fireEvent.click(
      await screen.findByRole("button", {
        name: "Abrir resultado Obra sintética",
      }),
    );
    expect(
      await screen.findByRole("region", { name: "Detalle del libro" }),
    ).toBeTruthy();
    expect(api.ListBooks).not.toHaveBeenCalled();
  });
  async function detail() {
    api.Dashboard.mockResolvedValue(overview(true));
    render(<App />);
    fireEvent.click(
      await screen.findByRole("button", { name: "Abrir Obra sintética" }),
    );
    await screen.findByRole("region", { name: "Detalle del libro" });
  }
  it("abre una serie con registros locales ordenados y libros adicionales de catálogo", async () => {
    const series = {
      id: "synthetic-series",
      name: "Serie sintética",
      position: 1,
    };
    const book = { ...sample, series };
    api.Dashboard.mockResolvedValue({
      ...overview(true),
      currentlyReading: [book],
    });
    api.ListSeriesBooks.mockResolvedValue([book]);
    catalog.Search.mockResolvedValue({
      works: [
        { id: "11", title: "Libro aún no agregado", authors: ["Autor A"] },
      ],
      warnings: [],
    });
    render(<App />);
    fireEvent.click(
      await screen.findByRole("button", { name: "Serie sintética · #1" }),
    );
    await waitFor(() =>
      expect(api.ListSeriesBooks).toHaveBeenCalledWith(series.id, "books"),
    );
    expect(
      await screen.findByLabelText("Estado de Obra sintética"),
    ).toBeTruthy();
    expect(
      screen.getByRole("heading", { name: /Serie sintética/ }),
    ).toBeTruthy();
    expect(
      await screen.findByRole("button", { name: /Libro aún no agregado/ }),
    ).toBeTruthy();
    expect(catalog.Search).toHaveBeenCalledWith(series.name);
  });
  it("abre resultados locales y remotos al pulsar el autor de un libro local", async () => {
    api.ListBooks.mockResolvedValue([sample]);
    catalog.Search.mockResolvedValue({
      works: [
        { id: "11", title: "Otro libro del autor", authors: ["Autor A"] },
      ],
      warnings: [],
    });
    await detail();
    fireEvent.click(screen.getByRole("button", { name: "Autor A" }));
    expect(
      await screen.findByRole("button", { name: /Otro libro del autor/ }),
    ).toBeTruthy();
    expect(api.ListBooks).toHaveBeenCalledWith("Autor A", "", "books");
    expect(catalog.Search).toHaveBeenCalledWith("Autor A");
    expect(
      screen.getByRole("button", { name: `Abrir ${sample.title}` }),
    ).toBeTruthy();
  });
  it("guarda notas desde la ficha sin reenviar el formulario completo", async () => {
    api.SaveBookNotes.mockResolvedValue({ ...sample, notes: "Nueva nota" });
    await detail();
    fireEvent.change(screen.getByLabelText("Notas personales del libro"), {
      target: { value: "Nueva nota" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Guardar notas" }));
    await waitFor(() =>
      expect(api.SaveBookNotes).toHaveBeenCalledWith(
        sample.id,
        "Nueva nota",
        "books",
      ),
    );
    expect((await screen.findByRole("status")).textContent).toContain(
      "Notas guardadas.",
    );
    expect(api.SaveBook).not.toHaveBeenCalled();
    expect(screen.queryByRole("region", { name: "Editar libro" })).toBeNull();
  });
  it("conserva el borrador de notas al marcar favorito y ante errores de guardado", async () => {
    api.SetBookFavorite.mockResolvedValue({ ...sample, favorite: true });
    api.SaveBookNotes.mockRejectedValue(new Error("Fallo sintético de notas."));
    await detail();
    const notes = screen.getByLabelText("Notas personales del libro");
    fireEvent.change(notes, { target: { value: "Borrador sin guardar" } });
    fireEvent.click(screen.getByRole("button", { name: "Favorito" }));
    await waitFor(() =>
      expect(
        screen
          .getByRole("button", { name: "Favorito" })
          .getAttribute("aria-pressed"),
      ).toBe("true"),
    );
    expect(notes).toHaveProperty("value", "Borrador sin guardar");
    fireEvent.click(screen.getByRole("button", { name: "Guardar notas" }));
    expect(await screen.findByRole("alert")).toHaveProperty(
      "textContent",
      "Fallo sintético de notas.",
    );
    expect(notes).toHaveProperty("value", "Borrador sin guardar");
  });
  it("marca propiedad mediante una operación específica", async () => {
    api.SetBookOwned.mockResolvedValue({
      ...sample,
      edition: { ...sample.edition, owned: true },
    });
    await detail();
    fireEvent.click(screen.getByRole("button", { name: "En propiedad" }));
    await waitFor(() =>
      expect(api.SetBookOwned).toHaveBeenCalledWith(sample.id, true, "books"),
    );
    expect(
      screen
        .getByRole("button", { name: "En propiedad" })
        .getAttribute("aria-pressed"),
    ).toBe("true");
    expect(api.SaveBook).not.toHaveBeenCalled();
  });
});

describe("guardar un libro manual", () => {
  it("conserva el formulario y muestra el error si el backend rechaza el guardado", async () => {
    api.SaveBook.mockRejectedValue(new Error("No se pudo guardar."));
    render(<App />);
    await screen.findByRole("heading", { name: "Tu biblioteca" });
    fireEvent.click(screen.getByRole("button", { name: "Añadir libro" }));
    fireEvent.change(screen.getByLabelText(/Título/), {
      target: { value: "Obra sintética" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Guardar libro" }));
    expect(await screen.findByRole("alert")).toHaveProperty(
      "textContent",
      "No se pudo guardar.",
    );
    expect(screen.getByLabelText(/Título/)).toHaveProperty(
      "value",
      "Obra sintética",
    );
    expect(
      screen.queryByText("Libro guardado en la biblioteca local."),
    ).toBeNull();
  });

  it("envía autoría y serie al backend y abre la ficha del registro confirmado", async () => {
    api.SaveBook.mockResolvedValue({
      ...sample,
      authors: ["Autor A", "Autora B"],
      series: { id: "series", name: "Serie sintética", position: 2 },
    });
    render(<App />);
    await screen.findByRole("heading", { name: "Tu biblioteca" });
    fireEvent.click(screen.getByRole("button", { name: "Añadir libro" }));
    fireEvent.change(screen.getByLabelText(/Título/), {
      target: { value: "Obra sintética" },
    });
    fireEvent.change(screen.getByLabelText(/^Autor o autores/), {
      target: { value: "Autor A\nAutora B" },
    });
    fireEvent.change(screen.getByLabelText("Serie"), {
      target: { value: "Serie sintética" },
    });
    fireEvent.change(screen.getByLabelText("N.º de orden"), {
      target: { value: "2" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Guardar libro" }));
    await waitFor(() => expect(api.SaveBook).toHaveBeenCalledOnce());
    expect(api.SaveBook.mock.calls[0][0]).toMatchObject({
      id: "",
      authors: ["Autor A", "Autora B"],
      seriesName: "Serie sintética",
      seriesPosition: 2,
    });
    expect(
      await screen.findByRole("heading", { name: "Obra sintética" }),
    ).toBeTruthy();
    expect(screen.queryByRole("button", { name: "Guardar libro" })).toBeNull();
  });
});

it("guarda clasificación manual del libro desde su ficha con el medio explícito", async () => {
  api.Dashboard.mockResolvedValue(overview(true));
  api.SetBookClassification.mockResolvedValue({
    ...sample,
    classification: { genres: ["Fantasía"], tags: [], hidden: [] },
    effectiveClassification: {
      genres: [
        {
          name: "Fantasía",
          origins: [{ provider: "manual", id: "fantasía", category: "genre" }],
        },
      ],
      tags: [],
    },
  });
  render(<App />);
  fireEvent.click(
    await screen.findByRole("button", { name: "Abrir Obra sintética" }),
  );
  const editor = await screen.findByRole("region", {
    name: "Géneros y etiquetas",
  });
  fireEvent.click(within(editor).getByText("Editar clasificación"));
  fireEvent.change(
    within(editor).getByLabelText("Nuevo término para Géneros y etiquetas"),
    { target: { value: "Fantasía" } },
  );
  fireEvent.click(within(editor).getByRole("button", { name: "Agregar" }));
  await waitFor(() =>
    expect(api.SetBookClassification).toHaveBeenCalledExactlyOnceWith(
      sample.id,
      "books",
      { genres: ["Fantasía"], tags: [], hidden: [] },
    ),
  );
  expect(await within(editor).findByText("Fantasía")).toBeTruthy();
});

describe("eliminar libros", () => {
  async function openDetail() {
    api.Dashboard.mockResolvedValue(overview(true));
    render(<App />);
    fireEvent.click(
      await screen.findByRole("button", { name: "Abrir Obra sintética" }),
    );
    await screen.findByRole("region", { name: "Detalle del libro" });
    fireEvent.click(screen.getByRole("button", { name: "Eliminar libro" }));
    return screen.findByRole("dialog");
  }
  it("no elimina al cancelar; solo envía la eliminación después de confirmar", async () => {
    api.DeleteBook.mockResolvedValue(undefined);
    const dialog = await openDetail();
    expect(api.DeleteBook).not.toHaveBeenCalled();
    fireEvent.click(within(dialog).getByRole("button", { name: "Cancelar" }));
    expect(screen.queryByRole("dialog")).toBeNull();
    expect(api.DeleteBook).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole("button", { name: "Eliminar libro" }));
    fireEvent.click(
      within(screen.getByRole("dialog")).getByRole("button", {
        name: "Eliminar definitivamente",
      }),
    );
    await waitFor(() =>
      expect(api.DeleteBook).toHaveBeenCalledWith(sample.id, "books"),
    );
    await waitFor(() => expect(screen.queryByRole("dialog")).toBeNull());
    expect(
      screen.queryByRole("region", { name: "Detalle del libro" }),
    ).toBeNull();
  });
  it("mantiene el detalle y la confirmación si el backend falla al eliminar", async () => {
    api.DeleteBook.mockRejectedValue(new Error("Error de eliminación."));
    const dialog = await openDetail();
    fireEvent.click(
      within(dialog).getByRole("button", { name: "Eliminar definitivamente" }),
    );
    expect(await within(dialog).findByRole("alert")).toHaveProperty(
      "textContent",
      "Error de eliminación.",
    );
    expect(
      screen.getByRole("region", { name: "Detalle del libro" }),
    ).toBeTruthy();
  });
});
