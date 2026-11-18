import { useEffect, useState } from "react";
import * as Catalog from "../../bindings/github.com/Santiparra/Trove/internal/desktop/catalog";
import type {
  Edition,
  Work,
} from "../../bindings/github.com/Santiparra/Trove/internal/catalog/models";
import BookSearch from "../books/BookSearch";
import { errorMessage } from "../books/presentation";

export default function EditionPicker({
  workID,
  disabled,
  onEdition,
  onWork,
  onClose,
  configured,
  localOnly = false,
}: {
  workID?: string;
  disabled: boolean;
  configured: boolean;
  localOnly?: boolean;
  onEdition: (id: string) => void;
  onWork: (work: Work) => void;
  onClose: () => void;
}) {
  const [editions, setEditions] = useState<Edition[]>([]);
  const [query, setQuery] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  useEffect(() => {
    let active = true;
    if (!workID || localOnly) {
      setEditions([]);
      setLoading(false);
      return;
    }
    setLoading(true);
    setError("");
    Catalog.Editions(workID)
      .then((items) => {
        if (active) setEditions(items);
      })
      .catch((reason) => {
        if (active) setError(errorMessage(reason));
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, [workID, localOnly]);
  return (
    <section className="edition-picker" aria-label="Ediciones">
      <div className="card-heading">
        <h2>Ediciones</h2>
        <button className="text-action" disabled={disabled} onClick={onClose}>
          Cerrar ediciones
        </button>
      </div>
      <BookSearch
        medium="books"
        value={query}
        disabled={disabled}
        configured={configured}
        localOnly={localOnly}
        onChange={setQuery}
        onBook={() => {}}
        onAll={() => {}}
        onWork={onWork}
        catalogOnly
      />
      {loading && <p role="status">Cargando ediciones…</p>}
      {error && <p role="alert">{error}</p>}
      <div className="edition-grid">
        {editions.map((edition) => (
          <button
            className="edition-choice"
            key={edition.id}
            disabled={disabled}
            onClick={() => onEdition(edition.id)}
          >
            <strong>{edition.title}</strong>
            <span>
              {edition.language || "Idioma no informado"} ·{" "}
              {edition.format || "Formato no informado"}
            </span>
            <span>
              {edition.publisher || "Editorial no informada"} ·{" "}
              {edition.pages ?? "?"} páginas
            </span>
          </button>
        ))}
      </div>
    </section>
  );
}
