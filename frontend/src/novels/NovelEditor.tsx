import { useState, type FormEvent } from "react";
import type {
  Input,
  Novel,
} from "../../bindings/github.com/Santiparra/Trove/internal/novels/models";
import type { Info } from "../../bindings/github.com/Santiparra/Trove/internal/desktop/models";

export default function NovelEditor({
  novel,
  info,
  busy,
  error,
  onSave,
  onCancel,
  onDirty,
}: {
  novel: Novel | null;
  info: Info;
  busy: boolean;
  error: string;
  onSave: (input: Input) => void;
  onCancel: () => void;
  onDirty: (value: boolean) => void;
}) {
  const [title, setTitle] = useState(novel?.title || "");
  const [authors, setAuthors] = useState(novel?.authors.join("; ") || "");
  const [status, setStatus] = useState(novel?.status || "to-read");
  const [total, setTotal] = useState(novel?.total?.toString() || "");
  const [completed, setCompleted] = useState(
    novel ? (novel.completed?.toString() ?? "") : "0",
  );
  const [notes, setNotes] = useState(novel?.notes || "");
  const [favorite, setFavorite] = useState(novel?.favorite || false);
  const [addedOn, setAddedOn] = useState(novel?.addedOn || "");
  function change(action: () => void) {
    action();
    onDirty(true);
  }
  function submit(event: FormEvent) {
    event.preventDefault();
    onSave({
      id: novel?.id || "",
      title,
      authors: authors
        .split(";")
        .map((value) => value.trim())
        .filter(Boolean),
      status,
      total: total === "" ? null : Number(total),
      completed: completed === "" ? null : Number(completed),
      notes,
      favorite,
      addedOn,
    });
  }
  return (
    <section className="editor-page" aria-label="Editor de novela">
      <button className="back-link" disabled={busy} onClick={onCancel}>
        Volver
      </button>
      <div className="page-heading">
        <h1>{novel ? "Editar novela" : "Añadir novela"}</h1>
      </div>
      <form className="form-card novel-form" onSubmit={submit}>
        <div className="novel-form-grid">
          <label>
            Título de la novela
            <input
              autoFocus
              required
              value={title}
              disabled={busy}
              onChange={(event) => change(() => setTitle(event.target.value))}
            />
          </label>
          <label>
            Autoría
            <input
              value={authors}
              disabled={busy}
              placeholder="Separar con punto y coma"
              onChange={(event) => change(() => setAuthors(event.target.value))}
            />
          </label>
          <label>
            Estado de la novela
            <select
              value={status}
              disabled={busy}
              onChange={(event) => change(() => setStatus(event.target.value))}
            >
              {info.statuses.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.value === "read" ? "Leída" : option.label}
                </option>
              ))}
            </select>
          </label>
          <label>
            Total de volúmenes
            <input
              type="number"
              min={1}
              max={10000}
              step={1}
              value={total}
              disabled={busy}
              placeholder="Desconocido"
              onChange={(event) => change(() => setTotal(event.target.value))}
            />
          </label>
          <label>
            Volúmenes completados
            <input
              type="number"
              min={0}
              max={total || 10000}
              step={1}
              value={completed}
              disabled={busy}
              placeholder="Cantidad desconocida"
              onChange={(event) =>
                change(() => setCompleted(event.target.value))
              }
            />
          </label>
          <label>
            Voy por el volumen
            <input
              type="number"
              min={1}
              max={total || 10000}
              step={1}
              aria-label="Voy por el volumen"
              value={
                completed === "" ||
                status === "read" ||
                (total !== "" && Number(completed) >= Number(total))
                  ? ""
                  : String(Number(completed) + 1)
              }
              disabled={
                busy ||
                status === "read" ||
                (total !== "" && Number(completed) >= Number(total))
              }
              onChange={(event) =>
                change(() =>
                  setCompleted(
                    event.target.value === ""
                      ? ""
                      : String(Number(event.target.value) - 1),
                  ),
                )
              }
            />
          </label>
          <label>
            Fecha de alta
            <input
              type="date"
              value={addedOn}
              disabled={busy}
              onChange={(event) => change(() => setAddedOn(event.target.value))}
            />
          </label>
          <label className="novel-checkbox">
            <input
              type="checkbox"
              checked={favorite}
              disabled={busy}
              onChange={(event) =>
                change(() => setFavorite(event.target.checked))
              }
            />{" "}
            Favorita
          </label>
        </div>
        <p className="management-hint">
          El progreso cuenta volúmenes terminados: 7/? deja el siguiente en 8.
          Podés dejar la cantidad vacía si no la conocés. Los terminados
          históricos conservan sus datos y no reciben fechas inventadas. Con
          total conocido, marcar Leída completa N/N.
        </p>
        <label>
          Notas de la novela
          <textarea
            rows={4}
            value={notes}
            disabled={busy}
            onChange={(event) => change(() => setNotes(event.target.value))}
          />
        </label>
        {error && (
          <p className="error-banner" role="alert">
            {error}
          </p>
        )}
        <div className="book-actions">
          <button className="primary" disabled={busy} type="submit">
            Guardar novela
          </button>
          <button
            className="quiet"
            disabled={busy}
            type="button"
            onClick={onCancel}
          >
            Cancelar
          </button>
        </div>
      </form>
    </section>
  );
}
