import { useState, type FormEvent } from "react";
import type {
  Volume,
  VolumeInput,
} from "../../bindings/github.com/Santiparra/Trove/internal/novels/models";
import type { Info } from "../../bindings/github.com/Santiparra/Trove/internal/desktop/models";

export default function VolumeEditor({
  volume,
  number,
  total,
  info,
  busy,
  error,
  onSave,
  onCancel,
  onDirty,
}: {
  volume: Volume | null;
  number: number;
  total: number | null;
  info: Info;
  busy: boolean;
  error: string;
  onSave: (input: VolumeInput) => void;
  onCancel: () => void;
  onDirty: (value: boolean) => void;
}) {
  const [input, setInput] = useState<VolumeInput>({
    id: volume?.id || "",
    number: volume?.number || number,
    title: volume?.title || "",
    description: volume?.description || "",
    language: volume?.edition.language || "",
    format: volume?.edition.format || "",
    identifier: volume?.edition.identifier || "",
    owned: volume?.edition.owned || false,
    favorite: volume?.favorite || false,
    notes: volume?.notes || "",
    startedOn: volume?.cycle.startedOn || "",
    finishedOn: volume?.cycle.finishedOn || "",
  });
  function change(values: Partial<VolumeInput>) {
    setInput((current) => ({ ...current, ...values }));
    onDirty(true);
  }
  function submit(event: FormEvent) {
    event.preventDefault();
    onSave(input);
  }
  return (
    <section className="editor-page" aria-label="Editor de volumen">
      <button className="back-link" onClick={onCancel} disabled={busy}>
        Volver a la novela
      </button>
      <div className="page-heading">
        <h1>
          {volume ? `Editar volumen ${volume.number}` : "Registrar volumen"}
        </h1>
      </div>
      <form className="form-card novel-form" onSubmit={submit}>
        <div className="novel-form-grid">
          <label>
            Número de volumen
            <input
              type="number"
              min={1}
              max={total || 10000}
              step={1}
              required
              disabled={busy || !!volume}
              value={input.number}
              onChange={(event) =>
                change({ number: Number(event.target.value) })
              }
            />
          </label>
          <label>
            Título del volumen
            <input
              placeholder="Opcional"
              value={input.title}
              disabled={busy}
              onChange={(event) => change({ title: event.target.value })}
            />
          </label>
          <label>
            Idioma
            <input
              value={input.language}
              disabled={busy}
              onChange={(event) => change({ language: event.target.value })}
            />
          </label>
          <label>
            Formato
            <select
              value={input.format}
              disabled={busy}
              onChange={(event) => change({ format: event.target.value })}
            >
              {info.formats.map((option) => (
                <option value={option.value} key={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </label>
          <label>
            ISBN / identificador
            <input
              value={input.identifier}
              disabled={busy}
              onChange={(event) => change({ identifier: event.target.value })}
            />
          </label>
          <label>
            Inicio del volumen
            <input
              type="date"
              value={input.startedOn}
              disabled={busy}
              onChange={(event) => change({ startedOn: event.target.value })}
            />
          </label>
          <label>
            Finalización del volumen
            <input
              type="date"
              value={input.finishedOn}
              disabled={busy}
              onChange={(event) => change({ finishedOn: event.target.value })}
            />
          </label>
          <label className="novel-checkbox">
            <input
              type="checkbox"
              checked={input.owned}
              disabled={busy}
              onChange={(event) => change({ owned: event.target.checked })}
            />{" "}
            En propiedad
          </label>
          <label className="novel-checkbox">
            <input
              type="checkbox"
              checked={input.favorite}
              disabled={busy}
              onChange={(event) => change({ favorite: event.target.checked })}
            />{" "}
            Volumen favorito
          </label>
        </div>
        <label>
          Sinopsis del volumen
          <textarea
            rows={4}
            value={input.description}
            disabled={busy}
            onChange={(event) => change({ description: event.target.value })}
          />
        </label>
        <label>
          Notas del volumen
          <textarea
            rows={4}
            value={input.notes}
            disabled={busy}
            onChange={(event) => change({ notes: event.target.value })}
          />
        </label>
        {error && (
          <p className="error-banner" role="alert">
            {error}
          </p>
        )}
        <div className="book-actions">
          <button className="primary" type="submit" disabled={busy}>
            Guardar volumen
          </button>
          <button
            className="quiet"
            type="button"
            disabled={busy}
            onClick={onCancel}
          >
            Cancelar
          </button>
        </div>
      </form>
    </section>
  );
}
