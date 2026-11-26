import { afterEach, expect, it, vi } from "vitest";
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import DescriptionText from "./DescriptionText";
const native = vi.hoisted(() => ({
  open: vi.fn().mockResolvedValue(undefined),
  copy: vi.fn().mockResolvedValue(undefined),
}));
vi.mock("@wailsio/runtime", () => ({
  Browser: { OpenURL: native.open },
  Clipboard: { SetText: native.copy },
}));
afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
});
it("mide desbordamiento real y actualiza Leer más al cambiar el espacio", () => {
  let height = 100;
  vi.spyOn(HTMLElement.prototype, "scrollHeight", "get").mockImplementation(
    () => height,
  );
  render(<DescriptionText text={"A long description. ".repeat(100)} />);
  expect(screen.queryByRole("button", { name: "Leer más" })).toBeNull();
  height = 400;
  fireEvent(window, new Event("resize"));
  fireEvent.click(screen.getByRole("button", { name: "Leer más" }));
  expect(
    screen
      .getByRole("button", { name: "Leer menos" })
      .getAttribute("aria-expanded"),
  ).toBe("true");
});
it("resalta URLs web y permite abrir o copiar sin seleccionar texto", async () => {
  render(
    <DescriptionText text="Source: https://example.com/book. javascript:alert(1)" />,
  );
  const link = screen.getByRole("link", { name: "https://example.com/book" });
  fireEvent.click(link);
  expect(native.open).toHaveBeenCalledWith("https://example.com/book");
  fireEvent.click(
    screen.getByRole("button", {
      name: "Copiar enlace https://example.com/book",
    }),
  );
  await screen.findByText("Copiado");
  expect(native.copy).toHaveBeenCalledWith("https://example.com/book");
  expect(screen.getAllByRole("link")).toHaveLength(1);
});
