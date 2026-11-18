import { afterEach, beforeEach, expect, it, vi } from "vitest";
import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
} from "@testing-library/react";
import ProviderHealth from "./ProviderHealth";

const api = vi.hoisted(() => ({ state: vi.fn(), check: vi.fn() }));
vi.mock(
  "../../bindings/github.com/Santiparra/Trove/internal/desktop/catalog",
  () => ({ ProviderHealth: api.state, CheckProviders: api.check }),
);
const healthy = (id: string, name: string) => ({
  id,
  name,
  state: "healthy",
  detail: "La API respondió correctamente.",
  checkedAt: "2026-01-01T10:00:00Z",
  latencyMS: 120,
  failures: 0,
  retryAt: "",
  checking: false,
});
const primary = healthy("hardcover", "Hardcover");
const secondary = healthy("openlibrary", "Open Library");
const report = {
  providers: [primary, secondary],
  checking: false,
  nextCheckAt: "",
};
beforeEach(() => {
  vi.resetAllMocks();
  api.state.mockResolvedValue(report);
});
afterEach(() => {
  cleanup();
  vi.useRealTimers();
});

it("muestra el resumen compacto y permite comprobar sin bloquear el resto de la app", async () => {
  api.check.mockResolvedValue({ ...report, checking: true });
  render(<ProviderHealth disabled={false} configured />);
  fireEvent.click(
    await screen.findByRole("button", { name: "Proveedores · OK" }),
  );
  expect(
    screen.getByRole("region", { name: "Salud de proveedores" }),
  ).toBeTruthy();
  expect(screen.getAllByText(/120 ms/)).toHaveLength(2);
  expect(api.check).not.toHaveBeenCalled();
  fireEvent.click(screen.getByRole("button", { name: "Comprobar ahora" }));
  await screen.findByRole("button", { name: "Proveedores · Comprobando…" });
  expect(screen.getByRole("button", { name: "Comprobando…" })).toHaveProperty(
    "disabled",
    true,
  );
  expect(api.check).toHaveBeenCalledOnce();
  fireEvent.keyDown(document, { key: "Escape" });
  expect(
    screen.queryByRole("region", { name: "Salud de proveedores" }),
  ).toBeNull();
});

it.each([
  ["unconfigured", "1 disponible", "Sin configurar"],
  ["access-denied", "Revisar acceso", "Acceso rechazado"],
  ["paused", "1 con problemas", "En pausa"],
])("distingue %s de una caída total", async (state, summary, label) => {
  api.state.mockResolvedValue({
    ...report,
    providers: [{ ...primary, state }, secondary],
  });
  render(
    <ProviderHealth disabled={false} configured={state !== "unconfigured"} />,
  );
  fireEvent.click(
    await screen.findByRole("button", { name: `Proveedores · ${summary}` }),
  );
  expect(screen.getByText(label)).toBeTruthy();
  expect(
    screen.queryByRole("button", { name: "Proveedores · Sin servicio" }),
  ).toBeNull();
});

it("refresca solo estado local para reflejar consultas normales y respeta la pausa manual", async () => {
  vi.useFakeTimers();
  render(<ProviderHealth disabled={false} configured />);
  await act(async () => {});
  api.state.mockResolvedValue({
    ...report,
    providers: [
      { ...primary, state: "paused" },
      { ...secondary, state: "paused" },
    ],
    nextCheckAt: new Date(Date.now() + 15000).toISOString(),
  });
  await act(() => vi.advanceTimersByTimeAsync(3000));
  fireEvent.click(
    screen.getByRole("button", { name: "Proveedores · Sin servicio" }),
  );
  expect(
    screen.getByRole("button", { name: "Comprobar ahora" }),
  ).toHaveProperty("disabled", true);
  expect(api.state).toHaveBeenCalledTimes(2);
  expect(api.check).not.toHaveBeenCalled();
});

it("un fallo del estado local no se presenta como proveedores saludables", async () => {
  api.state.mockRejectedValue(new Error("No se pudo leer el estado"));
  render(<ProviderHealth disabled={false} configured />);
  fireEvent.click(
    await screen.findByRole("button", { name: "Proveedores · Sin comprobar" }),
  );
  expect(await screen.findByRole("alert")).toHaveProperty(
    "textContent",
    "No se pudo leer el estado",
  );
});
