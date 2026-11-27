import { afterEach, beforeEach, expect, it, vi } from "vitest";
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from "@testing-library/react";
import { Info } from "../../bindings/github.com/Santiparra/Trove/internal/desktop/models";
import {
  Novel,
  Volume,
  ReadingRun,
} from "../../bindings/github.com/Santiparra/Trove/internal/novels/models";
import {
  Edition,
  Cycle,
} from "../../bindings/github.com/Santiparra/Trove/internal/books/models";
import NovelWorkspace from "./NovelWorkspace";

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
  RereadNovel: vi.fn(),
  RereadVolume: vi.fn(),
  SaveReading: vi.fn(),
  SaveRun: vi.fn(),
  SetRating: vi.fn(),
  SetClassification: vi.fn(),
}));
vi.mock(
  "../../bindings/github.com/Santiparra/Trove/internal/desktop/novels",
  () => api,
);
const info = new Info({
  statuses: [
    { value: "to-read", label: "Por leer" },
    { value: "up-next", label: "A continuación" },
    { value: "reading", label: "Leyendo" },
    { value: "read", label: "Leído" },
    { value: "paused", label: "Pausado" },
    { value: "did-not-finish", label: "Sin terminar" },
  ],
  formats: [
    { value: "", label: "Sin especificar" },
    { value: "digital", label: "Digital" },
  ],
});
const volume = (number: number, status = "read") =>
  new Volume({
    id: `volume-${number}`,
    number,
    title: "",
    edition: new Edition({ id: `edition-${number}` }),
    cycle: new Cycle({
      id: `cycle-${number}`,
      status,
      startedOn: "",
      finishedOn: "",
    }),
  });
const novel = (values: Partial<Novel> = {}) =>
  new Novel({
    id: "novel-1",
    title: "Novela sintética",
    authors: ["Autora"],
    status: "reading",
    completed: 0,
    total: null,
    notes: "Nota de novela",
    addedOn: "2026-01-01",
    volumes: [volume(1, "reading")],
    ...values,
  });
const onDirty = vi.fn();
const onNotice = vi.fn();
function mount() {
  return render(
    <NovelWorkspace
      info={info}
      onBusy={vi.fn()}
      onDirty={onDirty}
      canLeave={() => true}
      onNotice={onNotice}
    />,
  );
}
async function open(n: Novel) {
  api.List.mockResolvedValue([n]);
  api.Get.mockResolvedValue(n);
  mount();
  fireEvent.click(
    await screen.findByRole("button", { name: `Abrir novela ${n.title}` }),
  );
  await screen.findByRole("region", { name: "Detalle de novela" });
}
beforeEach(() => {
  vi.resetAllMocks();
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
    items: [],
    total: 0,
    page: 1,
    pages: 1,
    warnings: [],
  });
});
afterEach(cleanup);

it("edita etiquetas de volumen sin enviarlas a la clasificación de novela", async () => {
  const n = novel({
    status: "reading",
    completed: 0,
    volumes: [
      new Volume({
        ...volume(1),
        cycle: new Cycle({ id: "cycle-1", status: "reading" }),
      }),
    ],
    effectiveClassification: {
      genres: [
        {
          name: "Fantasy",
          origins: [{ provider: "ranobedb", id: "7", category: "genre" }],
        },
      ],
      tags: [],
    },
  });
  api.SetClassification.mockResolvedValue(
    new Novel({
      ...n,
      volumes: [
        new Volume({
          ...n.volumes[0],
          classification: { genres: [], tags: ["Cozy"], hidden: [] },
          effectiveClassification: {
            genres: [],
            tags: [
              {
                name: "Cozy",
                origins: [{ provider: "manual", id: "cozy", category: "tag" }],
              },
            ],
          },
        }),
      ],
    }),
  );
  await open(n);
  const editor = screen.getByRole("region", {
    name: "Clasificación del volumen 1",
  });
  fireEvent.click(
    within(editor).getByRole("button", { name: /Editar clasificación/ }),
  );
  fireEvent.change(
    within(editor).getByLabelText("Categoría para Clasificación del volumen 1"),
    { target: { value: "tag" } },
  );
  fireEvent.change(
    within(editor).getByLabelText(
      "Nuevo término para Clasificación del volumen 1",
    ),
    { target: { value: "Cozy" } },
  );
  fireEvent.click(within(editor).getByRole("button", { name: "Agregar" }));
  await waitFor(() =>
    expect(api.SetClassification).toHaveBeenCalledExactlyOnceWith(
      n.id,
      "volume-1",
      { genres: [], tags: ["Cozy"], hidden: [] },
    ),
  );
  expect(
    await within(editor).findByText("Cozy", {
      selector: ".classification-display .classification-chip",
    }),
  ).toBeTruthy();
  expect(
    within(
      screen.getByRole("region", { name: "Clasificación de novela" }),
    ).queryByText("Cozy"),
  ).toBeNull();
});

it("guarda el modo automático de novela y muestra la media recibida sin cuantizarla", async () => {
  const n = novel({ status: "reading", completed: 0, volumes: [volume(1)] });
  api.List.mockResolvedValue([n]);
  api.Get.mockResolvedValue(n);
  api.SetRating.mockResolvedValue(
    new Novel({ ...n, ratingMode: "auto", rating: 25.25 / 3 }),
  );
  mount();
  fireEvent.click(
    await screen.findByRole("button", { name: "Ver todas las novelas" }),
  );
  fireEvent.click(
    await screen.findByRole("button", { name: `Abrir novela ${n.title}` }),
  );
  const dropdown = await screen.findByRole("combobox", {
    name: "Puntaje de la novela",
  });
  const management = screen.getByRole("region", {
    name: "Seguimiento de novela",
  });
  expect(management.contains(dropdown)).toBe(false);
  expect(
    screen
      .getByRole("region", { name: "Gestión de novela" })
      .contains(dropdown),
  ).toBe(true);
  const volumeDropdown = screen.getByRole("combobox", {
    name: "Puntaje del volumen 1",
  });
  expect(management.contains(volumeDropdown)).toBe(false);
  expect(
    screen
      .getByRole("region", { name: "Gestión de volumen" })
      .contains(volumeDropdown),
  ).toBe(true);
  fireEvent.change(dropdown, {
    target: { value: "auto" },
  });
  expect(screen.queryByRole("button", { name: "Guardar puntaje" })).toBeNull();
  await waitFor(() =>
    expect(api.SetRating).toHaveBeenCalledExactlyOnceWith(n.id, "auto", null),
  );
  expect(
    await screen.findByLabelText("Resultado automático: Puntaje de la novela"),
  ).toHaveProperty("textContent", "8,42");
  expect(
    screen.queryByRole("combobox", {
      name: "Decimales de puntaje de la novela",
    }),
  ).toBeNull();
  expect(
    screen
      .getByRole("region", { name: "Gestión de novela" })
      .querySelector(".rating-select-label"),
  ).toBeNull();
});

it("prepara un recorrido con los volúmenes elegidos sin iniciarlos automáticamente", async () => {
  const n = novel({
    status: "read",
    completed: 3,
    total: 3,
    volumes: [volume(1), volume(2), volume(3)],
  });
  api.List.mockResolvedValue([n]);
  api.Get.mockResolvedValue(n);
  const run = new ReadingRun({
    id: "run",
    cycle: new Cycle({ id: "run", status: "up-next" }),
    volumeIDs: ["volume-2"],
    completed: 0,
  });
  api.RereadNovel.mockResolvedValue(
    new Novel({
      ...n,
      status: "up-next",
      completed: 0,
      activeRunID: "run",
      readings: [run],
    }),
  );
  mount();
  fireEvent.click(
    await screen.findByRole("button", { name: "Ver todas las novelas" }),
  );
  fireEvent.click(
    await screen.findByRole("button", { name: `Abrir novela ${n.title}` }),
  );
  await screen.findByRole("region", { name: "Detalle de novela" });
  fireEvent.click(screen.getByRole("tab", { name: "Recorridos" }));
  fireEvent.click(screen.getByRole("button", { name: "Releer novela" }));
  fireEvent.click(screen.getByLabelText("Volumen 1"));
  fireEvent.click(screen.getByLabelText("Volumen 3"));
  fireEvent.click(screen.getByRole("button", { name: "Crear recorrido" }));
  await waitFor(() =>
    expect(api.RereadNovel).toHaveBeenCalledExactlyOnceWith(n.id, ["volume-2"]),
  );
  await screen.findByRole("group", { name: "Progreso: 0/1" });
  expect(
    screen.getByRole("button", { name: "Empezar volumen 2" }),
  ).toHaveProperty("disabled", false);
  expect(api.StartVolume).not.toHaveBeenCalled();
  expect(api.FinishVolume).not.toHaveBeenCalled();
});

it("ofrece ambas relecturas desde sus historiales mientras la novela sigue en curso", async () => {
  const active = volume(3);
  active.cycle = new Cycle({
    id: "reading-3",
    status: "reading",
    startedOn: "2026-01-01",
  });
  const n = novel({
    status: "reading",
    completed: 2,
    total: null,
    volumes: [volume(1), volume(2), active],
  });
  api.List.mockResolvedValue([n]);
  api.Get.mockResolvedValue(n);
  const rereading = volume(1);
  rereading.cycle = new Cycle({
    id: "reread-1",
    status: "reading",
    startedOn: "2026-01-01",
  });
  api.RereadVolume.mockResolvedValue(
    new Novel({
      ...n,
      activeVolumeID: "volume-1",
      previousStatus: "reading",
      volumes: [rereading, volume(2), active],
    }),
  );
  mount();
  fireEvent.click(
    await screen.findByRole("button", { name: "Ver todas las novelas" }),
  );
  fireEvent.click(
    await screen.findByRole("button", { name: `Abrir novela ${n.title}` }),
  );
  fireEvent.click(await screen.findByRole("tab", { name: "Recorridos" }));
  const history = await screen.findByRole("region", {
    name: "Recorridos de novela",
  });
  expect(
    within(history).getByRole("button", { name: "Releer novela" }),
  ).toHaveProperty("disabled", false);
  fireEvent.click(screen.getByRole("button", { name: "Abrir volumen 1" }));
  fireEvent.click(screen.getByRole("tab", { name: "Lecturas del volumen" }));
  const volumeHistory = screen.getByRole("region", {
    name: "Lecturas del volumen 1",
  });
  const reread = within(volumeHistory).getByRole("button", {
    name: "Releer volumen",
  });
  expect(reread).toHaveProperty("disabled", false);
  fireEvent.click(reread);
  await waitFor(() =>
    expect(api.RereadVolume).toHaveBeenCalledExactlyOnceWith(n.id, "volume-1"),
  );
  expect(api.RereadNovel).not.toHaveBeenCalled();
});

it("lista una novela y su progreso, no una entrada por volumen", async () => {
  const n = novel({
    completed: 7,
    volumes: Array.from({ length: 7 }, (_, index) => volume(index + 1)),
  });
  api.List.mockResolvedValue([n]);
  mount();
  expect(
    await screen.findByRole("button", {
      name: "Abrir novela Novela sintética",
    }),
  ).toBeTruthy();
  expect(screen.getAllByRole("button", { name: /^Abrir novela/ })).toHaveLength(
    1,
  );
  expect(screen.getByText(/7\/\?/)).toBeTruthy();
  expect(screen.getByText(/Volumen 8/)).toBeTruthy();
});
it("abre en portal y Todas muestra realmente todos los estados desde el primer clic", async () => {
  const n = novel({ status: "to-read" });
  api.List.mockResolvedValue([n]);
  mount();
  await screen.findByRole("region", { name: "Portal de novelas" });
  expect(
    screen.queryByRole("button", { name: `Abrir novela ${n.title}` }),
  ).toBeNull();
  fireEvent.click(
    screen.getByRole("button", { name: "Ver todas las novelas" }),
  );
  await screen.findByRole("button", { name: `Abrir novela ${n.title}` });
  expect(
    screen.getByRole("button", { name: "Todas" }).classList.contains("active"),
  ).toBe(true);
});
it("ajusta progreso con una flecha y usa confirmación propia al borrar", async () => {
  const n = novel();
  api.SetCompleted.mockResolvedValue(
    novel({ completed: 1, status: "up-next" }),
  );
  api.Delete.mockResolvedValue(undefined);
  const confirm = vi.spyOn(window, "confirm");
  await open(n);
  fireEvent.click(
    screen.getByRole("button", { name: "Aumentar volúmenes completados" }),
  );
  await screen.findByRole("group", { name: "Progreso: 1/?" });
  expect(api.SetCompleted).toHaveBeenCalledExactlyOnceWith(n.id, 1, false);
  fireEvent.click(screen.getByRole("tab", { name: "Gestionar" }));
  fireEvent.click(screen.getByRole("button", { name: "Eliminar novela" }));
  expect(
    screen.getByRole("dialog", { name: "¿Eliminar esta novela?" }),
  ).toBeTruthy();
  expect(confirm).not.toHaveBeenCalled();
  expect(api.Delete).not.toHaveBeenCalled();
  fireEvent.click(
    screen.getByRole("button", { name: "Eliminar definitivamente" }),
  );
  await waitFor(() => expect(api.Delete).toHaveBeenCalledExactlyOnceWith(n.id));
});

it("el atajo Voy por el volumen X guarda X-1 y la creación no exige títulos de volumen", async () => {
  api.Save.mockResolvedValue(
    novel({
      status: "to-read",
      completed: 7,
      volumes: Array.from({ length: 7 }, (_, index) => volume(index + 1)),
    }),
  );
  mount();
  fireEvent.click(screen.getByRole("button", { name: "Añadir novela" }));
  fireEvent.change(screen.getByLabelText("Título de la novela"), {
    target: { value: "Novela sintética" },
  });
  fireEvent.change(screen.getByLabelText("Voy por el volumen"), {
    target: { value: "8" },
  });
  expect(screen.getByLabelText("Volúmenes completados")).toHaveProperty(
    "value",
    "7",
  );
  fireEvent.click(screen.getByRole("button", { name: "Guardar novela" }));
  await screen.findByRole("region", { name: "Detalle de novela" });
  expect(api.Save.mock.calls[0][0]).toMatchObject({
    title: "Novela sintética",
    completed: 7,
    total: null,
  });
  expect(
    screen.getByRole("button", { name: "Empezar volumen 8" }),
  ).toBeTruthy();
});

it("guarda Leída con cantidad desconocida sin dibujar volúmenes ficticios", async () => {
  api.Save.mockResolvedValue(
    novel({ status: "read", completed: null, total: null, volumes: [] }),
  );
  mount();
  fireEvent.click(screen.getByRole("button", { name: "Añadir novela" }));
  fireEvent.change(screen.getByLabelText("Título de la novela"), {
    target: { value: "Novela sintética" },
  });
  fireEvent.change(screen.getByLabelText("Estado de la novela"), {
    target: { value: "read" },
  });
  fireEvent.change(screen.getByLabelText("Volúmenes completados"), {
    target: { value: "" },
  });
  fireEvent.click(screen.getByRole("button", { name: "Guardar novela" }));
  await screen.findByRole("region", { name: "Detalle de novela" });
  expect(api.Save.mock.calls[0][0]).toMatchObject({
    status: "read",
    completed: null,
    total: null,
  });
  expect(screen.getByRole("group", { name: "Progreso: ?/?" })).toBeTruthy();
  expect(screen.queryByRole("button", { name: "Abrir volumen 1" })).toBeNull();
});

it("terminar deja el siguiente pendiente sin iniciarlo y conserva el borrador de notas", async () => {
  const n = novel();
  const updated = novel({
    status: "up-next",
    completed: 1,
    volumes: [
      volume(1),
      new Volume({
        ...volume(2, "to-read"),
        description: "Sinopsis del siguiente volumen",
      }),
    ],
  });
  api.FinishVolume.mockResolvedValue(updated);
  await open(n);
  fireEvent.click(screen.getByRole("tab", { name: "Notas" }));
  const notes = screen.getByLabelText("Notas de la novela");
  fireEvent.change(notes, {
    target: { value: "Borrador todavía sin guardar" },
  });
  fireEvent.click(screen.getByRole("button", { name: "Volumen terminado" }));
  await screen.findByRole("button", { name: "Empezar volumen 2" });
  expect(api.FinishVolume).toHaveBeenCalledExactlyOnceWith(n.id);
  expect(api.StartVolume).not.toHaveBeenCalled();
  expect(screen.getByText("Sinopsis del siguiente volumen")).toBeTruthy();
  expect(notes).toHaveProperty("value", "Borrador todavía sin guardar");
  expect(onDirty).toHaveBeenLastCalledWith(true);
});

it("al terminar N/N ofrece la acción explícita y no presenta un volumen N+1", async () => {
  const n = novel({
    completed: 1,
    total: 2,
    volumes: [volume(1), volume(2, "reading")],
  });
  api.FinishVolume.mockResolvedValue(
    novel({
      status: "read",
      completed: 2,
      total: 2,
      volumes: [volume(1), volume(2)],
    }),
  );
  await open(n);
  fireEvent.click(
    screen.getByRole("button", {
      name: "Terminar último volumen y marcar leída",
    }),
  );
  await waitFor(() =>
    expect(screen.getByLabelText(`Estado de ${n.title}`)).toHaveProperty(
      "value",
      "read",
    ),
  );
  expect(screen.queryByText(/Volumen 3/)).toBeNull();
  expect(
    screen.getByRole("button", { name: "Empezar volumen" }),
  ).toHaveProperty("disabled", true);
});

it("edita datos de un volumen con título opcional, conserva errores y no pierde el formulario", async () => {
  api.SaveVolume.mockRejectedValue(
    new Error("La finalización precede al inicio"),
  );
  await open(novel());
  fireEvent.click(screen.getByRole("button", { name: "Abrir volumen 1" }));
  fireEvent.click(screen.getByRole("tab", { name: "Gestionar" }));
  fireEvent.click(
    screen.getByRole("button", { name: "Editar datos del volumen 1" }),
  );
  expect(screen.getByLabelText("Posición de lectura")).toHaveProperty(
    "disabled",
    true,
  );
  expect(screen.getByLabelText("Título del volumen")).toHaveProperty(
    "required",
    false,
  );
  fireEvent.change(screen.getByLabelText("Notas del volumen"), {
    target: { value: "Nota por conservar" },
  });
  fireEvent.click(screen.getByRole("button", { name: "Guardar volumen" }));
  await screen.findByText("La finalización precede al inicio");
  expect(screen.getByLabelText("Notas del volumen")).toHaveProperty(
    "value",
    "Nota por conservar",
  );
  expect(api.SaveVolume.mock.calls[0]).toMatchObject([
    "novel-1",
    { id: "volume-1", number: 1, title: "", notes: "Nota por conservar" },
  ]);
});
