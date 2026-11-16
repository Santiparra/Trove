import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import App from "./App";

const api = vi.hoisted(() => ({
  Info: vi.fn(),
  ListBooks: vi.fn(),
  ListSeries: vi.fn(),
  SaveBook: vi.fn(),
  GetBook: vi.fn(),
}));
vi.mock(
  "../bindings/github.com/Santiparra/Trove/internal/desktop/library",
  () => api,
);

beforeEach(() => {
  api.Info.mockResolvedValue({
    profile: "test",
    dataDir: "synthetic/test",
    buildID: "synthetic",
    statuses: [{ value: "to-read", label: "Por leer" }],
    formats: [{ value: "", label: "Sin especificar" }],
  });
  api.ListBooks.mockResolvedValue([]);
  api.ListSeries.mockResolvedValue([]);
});
afterEach(cleanup);

describe("guardar un libro manual", () => {
  it("conserva el formulario y muestra el error si el backend rechaza el guardado", async () => {
    api.SaveBook.mockRejectedValue(new Error("No se pudo guardar."));
    render(<App />);
    await screen.findByText("Tu próxima historia empieza acá");
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

  it("envía autoría y serie al backend y muestra el registro confirmado, no un ID inventado", async () => {
    api.SaveBook.mockResolvedValue({
      id: "confirmed-id",
      title: "Obra sintética",
      authors: ["Autor A", "Autora B"],
      series: { id: "confirmed-series", name: "Serie sintética", position: 2 },
      edition: { id: "edition", language: "", format: "", identifier: "" },
      cycle: { id: "cycle", status: "to-read", startedOn: "", finishedOn: "" },
      notes: "",
      addedOn: "2026-01-01",
    });
    render(<App />);
    await screen.findByText("Tu próxima historia empieza acá");
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
