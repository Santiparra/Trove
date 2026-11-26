import { afterEach, expect, it, vi } from "vitest";
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import RatingEditor, { ExternalRating } from "./RatingEditor";
import { Rating } from "../../bindings/github.com/Santiparra/Trove/internal/catalog/models";
afterEach(cleanup);

it("guarda al elegir un cuarto de punto o Sin puntuar, sin formulario ni confirmación", async () => {
  const onSave = vi.fn().mockResolvedValue(true);
  const { rerender } = render(
    <RatingEditor value={8.25} busy={false} onSave={onSave} />,
  );
  const dropdown = screen.getByRole("combobox", { name: "Puntaje personal" });
  expect(dropdown).toHaveProperty("value", "8.25");
  expect(screen.getAllByRole("option")).toHaveLength(38);
  expect(screen.queryByRole("button", { name: "Guardar puntaje" })).toBeNull();
  fireEvent.change(dropdown, { target: { value: "9.75" } });
  await waitFor(() => expect(onSave).toHaveBeenCalledWith("manual", 9.75));
  rerender(<RatingEditor value={9.75} busy={false} onSave={onSave} />);
  expect(dropdown).toHaveProperty("value", "9.75");
  fireEvent.change(dropdown, { target: { value: "" } });
  await waitFor(() => expect(onSave).toHaveBeenCalledWith("", null));
});

it("guarda Auto inmediatamente y muestra el valor calculado por Go en la selección", async () => {
  const onSave = vi.fn().mockResolvedValue(true);
  const { rerender } = render(
    <RatingEditor
      value={null}
      mode=""
      allowAuto
      busy={false}
      onSave={onSave}
    />,
  );
  fireEvent.change(screen.getByRole("combobox"), { target: { value: "auto" } });
  await waitFor(() =>
    expect(onSave).toHaveBeenCalledExactlyOnceWith("auto", null),
  );
  rerender(
    <RatingEditor
      value={25.25 / 3}
      mode="auto"
      allowAuto
      busy={true}
      onSave={onSave}
    />,
  );
  expect(screen.getByRole("combobox")).toHaveProperty("value", "auto");
  expect(screen.getByRole("combobox")).toHaveProperty("disabled", true);
  expect(screen.getByRole("option", { name: "Auto · 8,42" })).toBeTruthy();
});

it("muestra solo nota grande y cantidad, conservando detalles externos al pasar el cursor", () => {
  const { rerender } = render(
    <ExternalRating
      provider="Hardcover · obra"
      rating={
        new Rating({
          provider: "hardcover",
          entity: "work",
          id: "10",
          value: 4.135,
          scale: 5,
          votes: 12,
          fetchedAt: "2026-01-01T00:00:00Z",
        })
      }
    />,
  );
  expect(screen.getByText("8,27")).toBeTruthy();
  expect(screen.getByText("12 valoraciones")).toBeTruthy();
  expect(screen.queryByText(/Original:/)).toBeNull();
  expect(
    screen.getByLabelText(/Puntaje externo de Hardcover/).getAttribute("title"),
  ).toContain("Original: 4.135/5");
  rerender(
    <ExternalRating
      provider="RanobeDB · novela"
      rating={
        new Rating({
          provider: "ranobedb",
          entity: "series",
          id: "10",
          value: 8.65,
          scale: 10,
          votes: 10,
          fetchedAt: "2026-01-01T00:00:00Z",
        })
      }
    />,
  );
  expect(screen.getByText("8,65")).toBeTruthy();
  rerender(<ExternalRating provider="RanobeDB · novela" rating={null} />);
  expect(screen.getByText("—")).toBeTruthy();
});
