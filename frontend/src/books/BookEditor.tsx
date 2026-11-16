import { useEffect, useState, type FormEvent } from "react";
import {
  Input,
  type Book,
  type Series,
} from "../../bindings/github.com/Santiparra/Trove/internal/books/models";
import type { Info } from "../../bindings/github.com/Santiparra/Trove/internal/desktop/models";
import Icon from "../components/Icon";

type Draft = {
  id: string;
  title: string;
  authors: string;
  seriesName: string;
  seriesPosition: string;
  language: string;
  format: string;
  identifier: string;
  status: string;
  notes: string;
  addedOn: string;
  startedOn: string;
  finishedOn: string;
};

function initialDraft(book: Book | null): Draft {
  return book
    ? {
        id: book.id,
        title: book.title,
        authors: book.authors.join("\n"),
        seriesName: book.series?.name ?? "",
        seriesPosition: book.series?.position?.toString() ?? "",
        language: book.edition.language,
        format: book.edition.format,
        identifier: book.edition.identifier,
        status: book.cycle.status,
        notes: book.notes,
        addedOn: book.addedOn,
        startedOn: book.cycle.startedOn,
        finishedOn: book.cycle.finishedOn,
      }
    : {
        id: "",
        title: "",
        authors: "",
        seriesName: "",
        seriesPosition: "",
        language: "",
        format: "",
        identifier: "",
        status: "to-read",
        notes: "",
        addedOn: "",
        startedOn: "",
        finishedOn: "",
      };
}

export default function BookEditor({
  book,
  info,
  series,
  busy,
  error,
  onCancel,
  onSave,
  onDirty,
}: {
  book: Book | null;
  info: Info;
  series: Series[];
  busy: boolean;
  error: string;
  onCancel: () => void;
  onSave: (input: Input) => Promise<void>;
  onDirty: (dirty: boolean) => void;
}) {
  const [initial] = useState(() => JSON.stringify(initialDraft(book)));
  const [draft, setDraft] = useState(() => initialDraft(book));
  useEffect(() => {
    onDirty(JSON.stringify(draft) !== initial);
  }, [draft, initial, onDirty]);
  const change = (field: keyof Draft, value: string) =>
    setDraft((current) => ({ ...current, [field]: value }));
  function save(event: FormEvent) {
    event.preventDefault();
    if (busy) return;
    void onSave(
      new Input({
        ...draft,
        authors: draft.authors.split("\n"),
        seriesPosition:
          draft.seriesPosition === "" ? null : Number(draft.seriesPosition),
      }),
    );
  }
  return (
    <section
      className="editor-page"
      aria-label={book ? "Editar libro" : "Añadir libro"}
    >
      <button className="back-link" onClick={onCancel} disabled={busy}>
        <Icon name="arrow" size={17} /> Volver
      </button>
      <header className="page-heading">
        <div>
          <p className="eyebrow">REGISTRO MANUAL · SIN PROVEEDOR</p>
          <h1>
            {book ? "Editar libro" : "Añadir libro"}
            <span className="heading-dot">.</span>
          </h1>
          <p>
            Solo el título es obligatorio. El resto puede completarse después.
          </p>
        </div>
      </header>
      <form onSubmit={save}>
        <div className="form-grid">
          <fieldset disabled={busy} className="form-card">
            <legend>Identificación</legend>
            <label>
              Título <span className="required">*</span>
              <input
                autoFocus
                required
                value={draft.title}
                onChange={(e) => change("title", e.target.value)}
                placeholder="Título del libro"
              />
            </label>
            <label>
              Autor o autores
              <textarea
                rows={3}
                value={draft.authors}
                onChange={(e) => change("authors", e.target.value)}
                placeholder="Un nombre por línea"
              />
              <small>Un autor por línea. Podés completarlo después.</small>
            </label>
            <div className="form-row series-fields">
              <label>
                Serie
                <input
                  list="series-options"
                  value={draft.seriesName}
                  onChange={(e) => change("seriesName", e.target.value)}
                  placeholder="Sin serie"
                />
                <datalist id="series-options">
                  {series.map((group) => (
                    <option key={group.id} value={group.name} />
                  ))}
                </datalist>
              </label>
              <label>
                N.º de orden
                <input
                  type="number"
                  min="0"
                  step="any"
                  value={draft.seriesPosition}
                  onChange={(e) => change("seriesPosition", e.target.value)}
                  placeholder="Ej. 1"
                />
              </label>
            </div>
            <small>
              Elegí una serie existente o escribí un nombre para vincular los
              libros.
            </small>
          </fieldset>

          <fieldset disabled={busy} className="form-card">
            <legend>
              Edición <span>OPCIONAL</span>
            </legend>
            <div className="form-row">
              <label>
                Idioma
                <input
                  value={draft.language}
                  onChange={(e) => change("language", e.target.value)}
                  placeholder="Ej. Inglés"
                />
              </label>
              <label>
                Formato
                <select
                  value={draft.format}
                  onChange={(e) => change("format", e.target.value)}
                >
                  {info.formats.map((option) => (
                    <option key={option.value} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </select>
              </label>
            </div>
            <label>
              ISBN / identificador
              <input
                value={draft.identifier}
                onChange={(e) => change("identifier", e.target.value)}
                placeholder="No es necesario para guardar"
              />
            </label>
            <p className="field-help">
              Esta edición es local. No hace falta encontrarla en un catálogo
              externo para registrarla.
            </p>
          </fieldset>

          <fieldset disabled={busy} className="form-card">
            <legend>Tu lectura</legend>
            <label>
              Estado
              <select
                value={draft.status}
                onChange={(e) => change("status", e.target.value)}
              >
                {info.statuses.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>
            </label>
            <div className="form-row">
              <label>
                Inicio
                <input
                  type="date"
                  value={draft.startedOn}
                  onChange={(e) => change("startedOn", e.target.value)}
                />
              </label>
              <label>
                Finalización
                <input
                  type="date"
                  value={draft.finishedOn}
                  onChange={(e) => change("finishedOn", e.target.value)}
                />
              </label>
            </div>
            <label>
              Añadido el
              <input
                type="date"
                value={draft.addedOn}
                onChange={(e) => change("addedOn", e.target.value)}
              />
              <small>
                {book
                  ? "Fecha de incorporación a tu biblioteca."
                  : "Si lo dejás vacío, se registra la fecha de hoy."}
              </small>
            </label>
          </fieldset>

          <fieldset disabled={busy} className="form-card">
            <legend>Notas personales</legend>
            <label className="notes-label">
              Tus impresiones
              <textarea
                rows={9}
                value={draft.notes}
                onChange={(e) => change("notes", e.target.value)}
                placeholder="Qué te dejó esta historia, ideas, recordatorios…"
              />
            </label>
          </fieldset>
        </div>
        {error && (
          <div className="error-banner" role="alert">
            {error}
          </div>
        )}
        <div className="editor-actions">
          <span>Se guarda únicamente en esta biblioteca local.</span>
          <button
            type="button"
            className="quiet"
            onClick={onCancel}
            disabled={busy}
          >
            Cancelar
          </button>
          <button type="submit" className="primary" disabled={busy}>
            {busy ? "Guardando…" : "Guardar libro"}
          </button>
        </div>
      </form>
    </section>
  );
}
