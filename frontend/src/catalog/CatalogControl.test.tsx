import { afterEach, expect, it, vi } from "vitest";
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import CatalogControl from "./CatalogControl";
const open = vi.hoisted(() => vi.fn().mockResolvedValue(undefined));
vi.mock("@wailsio/runtime", () => ({ Browser: { OpenURL: open } }));
vi.mock(
  "../../bindings/github.com/Santiparra/Trove/internal/desktop/catalog",
  () => ({}),
);
afterEach(cleanup);
it("destaca claves pendientes y abre la página oficial sin consultar ni revelar una clave", () => {
  render(
    <CatalogControl
      state={{ provider: "hardcover", configured: false }}
      disabled={false}
      onState={vi.fn()}
    />,
  );
  const indicator = screen.getByRole("button", { name: "Claves · 0" });
  expect(indicator.classList.contains("keys-pending")).toBe(true);
  fireEvent.click(indicator);
  fireEvent.click(
    screen.getByRole("button", { name: /Obtener clave en Hardcover/ }),
  );
  expect(open).toHaveBeenCalledExactlyOnceWith(
    "https://hardcover.app/account/api",
  );
  expect(screen.getByLabelText("Clave de catálogo")).toHaveProperty(
    "value",
    "",
  );
});
