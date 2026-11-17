import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import BackupPanel from "./BackupPanel";

const api = vi.hoisted(() => ({
  State: vi.fn(),
  ChooseFolder: vi.fn(),
  BackupNow: vi.fn(),
  ChooseRestore: vi.fn(),
  CancelRestore: vi.fn(),
  ConfirmRestore: vi.fn(),
}));
vi.mock(
  "../../bindings/github.com/Santiparra/Trove/internal/desktop/backups",
  () => api,
);
const state = {
  profile: "test",
  folder: "synthetic/backups/test",
  phase: "current",
  pending: false,
  lastSavedAt: "2026-01-01T00:00:00Z",
  generation: "synthetic-generation",
  lastError: "",
};
const preview = {
  token: "synthetic-token",
  source: "synthetic/trove.json",
  kind: "JSON",
  createdAt: "2026-01-01T00:00:00Z",
  generation: "synthetic-generation",
  profile: "test",
  dataDir: "synthetic/test",
  works: 2,
  editions: 2,
  cycles: 2,
};
beforeEach(() => {
  vi.resetAllMocks();
  api.State.mockResolvedValue(state);
  api.ChooseRestore.mockResolvedValue(preview);
  api.CancelRestore.mockResolvedValue(undefined);
});
afterEach(cleanup);

describe("restauración confirmada", () => {
  it("validar no restaura y el reemplazo exige una confirmación explícita", async () => {
    const restored = vi.fn();
    api.ConfirmRestore.mockResolvedValue({
      safetyCopy: "synthetic/before-restore.sqlite",
      works: 2,
    });
    render(
      <BackupPanel
        state={state}
        onState={vi.fn()}
        onClose={vi.fn()}
        onRestored={restored}
      />,
    );
    fireEvent.click(
      screen.getByRole("button", {
        name: "Seleccionar respaldo para restaurar",
      }),
    );
    const confirm = await screen.findByRole("button", {
      name: "Restaurar y reemplazar biblioteca",
    });
    expect(confirm).toHaveProperty("disabled", true);
    expect(api.ConfirmRestore).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole("checkbox"));
    fireEvent.click(confirm);
    await waitFor(() =>
      expect(api.ConfirmRestore).toHaveBeenCalledWith(preview.token),
    );
    await waitFor(() => expect(restored).toHaveBeenCalledOnce());
    expect(screen.getByText("synthetic/before-restore.sqlite")).toBeTruthy();
  });
  it("cancelar la preparación descarta su token y no reemplaza la biblioteca", async () => {
    render(
      <BackupPanel
        state={state}
        onState={vi.fn()}
        onClose={vi.fn()}
        onRestored={vi.fn()}
      />,
    );
    fireEvent.click(
      screen.getByRole("button", {
        name: "Seleccionar respaldo para restaurar",
      }),
    );
    fireEvent.click(
      await screen.findByRole("button", { name: "Cancelar preparación" }),
    );
    await waitFor(() =>
      expect(api.CancelRestore).toHaveBeenCalledWith(preview.token),
    );
    expect(api.ConfirmRestore).not.toHaveBeenCalled();
    expect(screen.queryByRole("checkbox")).toBeNull();
  });
  it("un error de restauración no muestra éxito ni descarta el error", async () => {
    const restored = vi.fn();
    api.ConfirmRestore.mockRejectedValue(
      new Error("Error sintético de reemplazo."),
    );
    render(
      <BackupPanel
        state={state}
        onState={vi.fn()}
        onClose={vi.fn()}
        onRestored={restored}
      />,
    );
    fireEvent.click(
      screen.getByRole("button", {
        name: "Seleccionar respaldo para restaurar",
      }),
    );
    fireEvent.click(await screen.findByRole("checkbox"));
    fireEvent.click(
      screen.getByRole("button", { name: "Restaurar y reemplazar biblioteca" }),
    );
    expect(await screen.findByRole("alert")).toHaveProperty(
      "textContent",
      "Error sintético de reemplazo.",
    );
    expect(restored).not.toHaveBeenCalled();
  });
});
