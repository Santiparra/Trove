import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from "@testing-library/react";
import App from "./App";
import * as BackupService from "../bindings/github.com/Santiparra/Trove/internal/desktop/backups";

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
      expect(api.ListBooks).toHaveBeenCalledWith("", "reading"),
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
    expect(api.ChangeBookStatus).toHaveBeenCalledWith(sample.id, "read");
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
      expect(api.SearchBooks).toHaveBeenCalledWith("sintética"),
    );
    expect(api.ListBooks).not.toHaveBeenCalled();
    expect(screen.getByRole("heading", { name: "Tu biblioteca" })).toBeTruthy();
    fireEvent.click(
      await screen.findByRole("button", {
        name: "Ver todos los resultados (0)",
      }),
    );
    await waitFor(() =>
      expect(api.ListBooks).toHaveBeenCalledWith("sintética", ""),
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
      expect(api.ListBooks).toHaveBeenCalledWith("", "up-next"),
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
  it("abre una página local de serie usando su identidad y conserva controles de lectura", async () => {
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
    render(<App />);
    fireEvent.click(
      await screen.findByRole("button", { name: "Serie sintética · #1" }),
    );
    await waitFor(() =>
      expect(api.ListSeriesBooks).toHaveBeenCalledWith(series.id),
    );
    expect(
      await screen.findByLabelText("Estado de Obra sintética"),
    ).toBeTruthy();
    expect(
      screen.getByRole("heading", { name: /Serie sintética/ }),
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
      expect(api.SaveBookNotes).toHaveBeenCalledWith(sample.id, "Nueva nota"),
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
      expect(api.SetBookOwned).toHaveBeenCalledWith(sample.id, true),
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
    await waitFor(() => expect(api.DeleteBook).toHaveBeenCalledWith(sample.id));
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
