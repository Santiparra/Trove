import { afterEach, expect, it, vi } from "vitest";
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import ClassificationEditor from "./ClassificationEditor";
import {
  Input,
  Set as ClassificationSet,
} from "../../bindings/github.com/Santiparra/Trove/internal/classification/models";
afterEach(cleanup);

it("agrega términos manuales sin reemplazar preferencias y conserva el borrador si falla", async () => {
  const onSave = vi.fn().mockResolvedValue(false),
    onDirty = vi.fn();
  render(
    <ClassificationEditor
      title="Libro"
      manual={
        new Input({
          genres: ["Fantasy"],
          hidden: [{ kind: "tag", provider: "hardcover", id: "5" }],
        })
      }
      busy={false}
      onSave={onSave}
      onDirty={onDirty}
    />,
  );
  fireEvent.click(screen.getByText("Editar clasificación"));
  fireEvent.change(screen.getByLabelText("Categoría para Libro"), {
    target: { value: "tag" },
  });
  fireEvent.change(screen.getByLabelText("Nuevo término para Libro"), {
    target: { value: "Cozy" },
  });
  expect(onDirty).toHaveBeenLastCalledWith(true);
  fireEvent.click(screen.getByRole("button", { name: "Agregar" }));
  await waitFor(() =>
    expect(onSave).toHaveBeenCalledWith({
      genres: ["Fantasy"],
      tags: ["Cozy"],
      hidden: [{ kind: "tag", provider: "hardcover", id: "5" }],
    }),
  );
  expect(screen.getByLabelText("Nuevo término para Libro")).toHaveProperty(
    "value",
    "Cozy",
  );
  onSave.mockResolvedValue(true);
  fireEvent.click(screen.getByRole("button", { name: "Agregar" }));
  await waitFor(() =>
    expect(screen.getByLabelText("Nuevo término para Libro")).toHaveProperty(
      "value",
      "",
    ),
  );
  expect(onDirty).toHaveBeenLastCalledWith(false);
});

it("oculta por identidad externa al quitar un término combinado y permite restaurarlo", async () => {
  const onSave = vi.fn().mockResolvedValue(true);
  const effective = new ClassificationSet({
    genres: [
      {
        name: "Fantasy",
        origins: [
          { provider: "manual", id: "fantasy", category: "genre" },
          { provider: "hardcover", id: "4", category: "genre" },
        ],
      },
    ],
  });
  const { rerender } = render(
    <ClassificationEditor
      title="Libro"
      manual={new Input({ genres: ["Fantasy"] })}
      effective={effective}
      onSave={onSave}
    />,
  );
  fireEvent.click(screen.getByText("Editar clasificación"));
  fireEvent.click(
    await screen.findByRole("button", {
      name: "Quitar género Fantasy de Libro",
    }),
  );
  await waitFor(() =>
    expect(onSave).toHaveBeenCalledWith({
      genres: [],
      tags: [],
      hidden: [{ kind: "genre", provider: "hardcover", id: "4" }],
    }),
  );
  rerender(
    <ClassificationEditor
      title="Libro"
      manual={
        new Input({
          hidden: [{ kind: "genre", provider: "hardcover", id: "4" }],
        })
      }
      onSave={onSave}
    />,
  );
  fireEvent.click(
    screen.getByRole("button", {
      name: "Restaurar términos del proveedor (1)",
    }),
  );
  await waitFor(() =>
    expect(onSave).toHaveBeenLastCalledWith({
      genres: [],
      tags: [],
      hidden: [],
    }),
  );
});

it("presenta procedencia sin ofrecer edición sobre una ficha remota", () => {
  render(
    <ClassificationEditor
      title="Catálogo"
      effective={
        new ClassificationSet({
          tags: [
            {
              name: "Adventurous",
              origins: [{ provider: "hardcover", id: "5", category: "mood" }],
            },
          ],
        })
      }
    />,
  );
  expect(screen.getByText("Adventurous").getAttribute("title")).toContain(
    "hardcover · mood · 5",
  );
  expect(screen.queryByText("Editar clasificación")).toBeNull();
  expect(screen.queryByRole("button")).toBeNull();
});
