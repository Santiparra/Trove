import { afterEach, expect, it, vi } from "vitest";
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { Cycle } from "../../bindings/github.com/Santiparra/Trove/internal/books/models";
import { Info } from "../../bindings/github.com/Santiparra/Trove/internal/desktop/models";
import ReadingHistory from "./ReadingHistory";

afterEach(cleanup);

it("conserva el valor automático completo al corregir fechas y lo muestra con dos decimales", async () => {
  const rating = 25.25 / 3;
  const c = new Cycle({
    id: "auto",
    status: "read",
    rating,
    ratingAutomatic: true,
  });
  const onSave = vi.fn().mockResolvedValue(true);
  render(
    <ReadingHistory
      current={c}
      info={info}
      busy={false}
      onSave={onSave}
      onDirty={vi.fn()}
      allowAdd={false}
    />,
  );
  expand();
  fireEvent.click(
    screen.getByRole("button", { name: "Corregir lectura actual" }),
  );
  expect(screen.getByLabelText("Puntaje de esta lectura")).toHaveProperty(
    "disabled",
    true,
  );
  expect(screen.getByLabelText("Puntaje de esta lectura")).toHaveProperty(
    "value",
    "8.42",
  );
  fireEvent.change(screen.getByLabelText("Fin de esta lectura"), {
    target: { value: "2020-01-02" },
  });
  fireEvent.click(screen.getByRole("button", { name: "Guardar lectura" }));
  await waitFor(() =>
    expect(onSave).toHaveBeenCalledWith(
      expect.objectContaining({ id: "auto", rating, finishedOn: "2020-01-02" }),
    ),
  );
});
const info = new Info({
  statuses: [
    { value: "read", label: "Leído" },
    { value: "reading", label: "Leyendo" },
    { value: "did-not-finish", label: "Sin terminar" },
  ],
  formats: [
    { value: "", label: "Sin especificar" },
    { value: "digital", label: "Digital" },
  ],
});
const current = new Cycle({
  id: "current",
  status: "read",
  startedOn: "2022-01-01",
  finishedOn: "2022-01-03",
});
function expand() {
  fireEvent.click(screen.getByText(/Historial ·/));
}

it("agrega una lectura histórica con fechas desconocidas sin iniciar una relectura", async () => {
  const onSave = vi.fn().mockResolvedValue(true),
    onDirty = vi.fn(),
    onReread = vi.fn();
  render(
    <ReadingHistory
      current={current}
      info={info}
      busy={false}
      onSave={onSave}
      onDirty={onDirty}
      onReread={onReread}
    />,
  );
  expand();
  fireEvent.click(
    screen.getByRole("button", { name: "Agregar lectura pasada" }),
  );
  fireEvent.change(screen.getByLabelText("Puntaje de esta lectura"), {
    target: { value: "8.25" },
  });
  expect(onDirty).toHaveBeenLastCalledWith(true);
  fireEvent.click(screen.getByRole("button", { name: "Guardar lectura" }));
  await waitFor(() =>
    expect(onSave).toHaveBeenCalledWith({
      id: "",
      status: "read",
      startedOn: "",
      finishedOn: "",
      rating: 8.25,
      edition: null,
    }),
  );
  await waitFor(() =>
    expect(
      screen.queryByRole("button", { name: "Guardar lectura" }),
    ).toBeNull(),
  );
  expect(onReread).not.toHaveBeenCalled();
  expect(onDirty).toHaveBeenLastCalledWith(false);
});

it("corrige un ciclo por su identidad y conserva el borrador si falla el guardado", async () => {
  const onSave = vi.fn().mockResolvedValue(false),
    onDirty = vi.fn(),
    onReread = vi.fn();
  const old = new Cycle({
    id: "older",
    status: "read",
    startedOn: "",
    finishedOn: "2020-01-01",
    rating: 7.5,
  });
  render(
    <ReadingHistory
      current={current}
      history={[old]}
      info={info}
      busy={false}
      onSave={onSave}
      onDirty={onDirty}
      onReread={onReread}
    />,
  );
  expand();
  fireEvent.click(
    screen.getByRole("button", { name: "Corregir lectura older" }),
  );
  fireEvent.change(screen.getByLabelText("Fin de esta lectura"), {
    target: { value: "2020-01-02" },
  });
  fireEvent.click(screen.getByRole("button", { name: "Guardar lectura" }));
  await waitFor(() =>
    expect(onSave).toHaveBeenCalledWith(
      expect.objectContaining({
        id: "older",
        finishedOn: "2020-01-02",
        rating: 7.5,
      }),
    ),
  );
  expect(screen.getByLabelText("Fin de esta lectura")).toHaveProperty(
    "value",
    "2020-01-02",
  );
  expect(onDirty).toHaveBeenLastCalledWith(true);
  expect(onReread).not.toHaveBeenCalled();
});
