import { useState, type FormEvent } from "react";
import type {
  Input,
  Novel,
} from "../../bindings/github.com/Santiparra/Trove/internal/novels/models";
import type { Info } from "../../bindings/github.com/Santiparra/Trove/internal/desktop/models";
import { volumeLabel } from "./presentation";

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
  const [progressBookID, setProgressBookID] = useState("");
  const [progressVolumeID, setProgressVolumeID] = useState("");
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
      progressBookID,
      progressVolumeID,
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
                change(() => {
                  setCompleted(event.target.value);
                  setProgressBookID("");
                  setProgressVolumeID("");
                })
              }
            />
          </label>
          <label>
            Voy por el volumen
            {novel?.catalog ? (
              <select
                aria-label="Voy por el volumen"
                disabled={busy || status === "read"}
                value={
                  progressBookID ||
                  novel.catalog.books.find(
                    (ref) => ref.number === Number(completed) + 1,
                  )?.id ||
                  ""
                }
                onChange={(event) =>
                  change(() => {
                    const ref = novel.catalog!.books.find(
                      (book) => book.id === event.target.value,
                    );
                    setProgressBookID(ref?.id || "");
                    setProgressVolumeID("");
                    setCompleted(ref ? String(ref.number - 1) : "");
                  })
                }
              >
                <option value="">Sin volumen identificado</option>
                {novel.catalog.books.map((ref) => (
                  <option value={ref.id} key={ref.id}>
                    {volumeLabel(novel, ref.number)} · {ref.title}
                  </option>
                ))}
              </select>
            ) : novel && novel.volumes.some((volume) => volume.label) ? (
              <select
                aria-label="Voy por el volumen"
                disabled={busy || status === "read"}
                value={
                  progressVolumeID ||
                  novel.volumes.find(
                    (volume) => volume.number === Number(completed) + 1,
                  )?.id ||
                  ""
                }
                onChange={(event) =>
                  change(() => {
                    const volume = novel.volumes.find(
                      (volume) => volume.id === event.target.value,
                    );
                    setProgressVolumeID(volume?.id || "");
                    setProgressBookID("");
                    setCompleted(volume ? String(volume.number - 1) : "");
                  })
                }
              >
                <option value="">Sin volumen identificado</option>
                {novel.volumes.map((volume) => (
                  <option value={volume.id} key={volume.id}>
                    {volumeLabel(novel, volume.number)} · {volume.title}
                  </option>
                ))}
              </select>
            ) : (
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
            )}
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
          El progreso cuenta publicaciones terminadas, incluidos decimales. La
          etiqueta del volumen se muestra aparte; con catálogo el atajo
          identifica el libro, no calcula X−1 desde su nombre. Podés dejar la
          cantidad vacía si no la conocés. Los terminados históricos conservan
          sus datos y no reciben fechas inventadas. Con total conocido, marcar
          Leída completa N/N.
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
