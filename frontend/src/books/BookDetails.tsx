import { useEffect, useState, type ReactNode } from "react";
import type {
  Book,
  Series,
} from "../../bindings/github.com/Santiparra/Trove/internal/books/models";
import type { Info } from "../../bindings/github.com/Santiparra/Trove/internal/desktop/models";
import Icon from "../components/Icon";
import StatusSelect from "./StatusSelect";
import { dateLabel } from "./presentation";
import { readingLabels } from "./readingMedia";
import BookProfile from "./BookProfile";

export default function BookDetails({
  book,
  info,
  busy,
  onBack,
  onEdit,
  onDelete,
  onStatus,
  onSeries,
  onFavorite,
  onOwned,
  onNotes,
  onDirty,
  onCatalog,
  onRefresh,
  onUnlink,
  editions,
}: {
  book: Book;
  info: Info;
  busy: boolean;
  onBack: () => void;
  onEdit: () => void;
  onDelete: (book: Book) => void;
  onStatus: (book: Book, status: string) => void;
  onSeries: (series: Series) => void;
  onFavorite: (book: Book, value: boolean) => void;
  onOwned: (book: Book, value: boolean) => void;
  onNotes: (book: Book, notes: string) => Promise<void>;
  onDirty: (value: boolean) => void;
  onCatalog: () => void;
  onRefresh: () => void;
  onUnlink: () => void;
  editions?: ReactNode;
}) {
  const [notes, setNotes] = useState(book.notes);
  const labels = readingLabels(book.medium);
  useEffect(() => setNotes(book.notes), [book.id, book.notes]);
  useEffect(() => onDirty(notes !== book.notes), [notes, book.notes, onDirty]);
  return (
    <BookProfile
      book={book}
      busy={busy}
      onBack={onBack}
      onSeries={onSeries}
      management={
        <>
          <StatusSelect
            book={book}
            options={info.statuses}
            disabled={busy}
            onChange={onStatus}
          />
          <div className="personal-flags">
            <button
              className={`flag-button ${book.favorite ? "is-active" : ""}`}
              aria-pressed={book.favorite}
              disabled={busy}
              onClick={() => onFavorite(book, !book.favorite)}
            >
              <Icon name="star" size={16} /> Favorito
            </button>
            <button
              className={`flag-button ${book.edition.owned ? "is-active" : ""}`}
              aria-pressed={book.edition.owned}
              disabled={busy}
              onClick={() => onOwned(book, !book.edition.owned)}
            >
              <Icon name="check" size={16} /> En propiedad
            </button>
          </div>
          <button className="secondary" onClick={onEdit} disabled={busy}>
            <Icon name="edit" size={16} /> {labels.edit}
          </button>
          {book.medium === "books" && (
            <>
              <button className="quiet" disabled={busy} onClick={onCatalog}>
                {book.edition.catalog
                  ? "Cambiar edición vinculada"
                  : "Vincular edición de catálogo"}
              </button>
              {book.edition.catalog && (
                <>
                  <button className="quiet" disabled={busy} onClick={onRefresh}>
                    Refrescar metadata
                  </button>
                  <button className="quiet" disabled={busy} onClick={onUnlink}>
                    Desvincular catálogo
                  </button>
                </>
              )}
            </>
          )}
          <button
            className="quiet danger-text"
            disabled={busy}
            onClick={() => onDelete(book)}
          >
            <Icon name="trash" size={16} /> Eliminar {labels.item}
          </button>
          <dl className="profile-facts personal-dates">
            <dt>Añadido</dt>
            <dd>{dateLabel(book.addedOn)}</dd>
            <dt>Inicio</dt>
            <dd>{dateLabel(book.cycle.startedOn)}</dd>
            <dt>Finalización</dt>
            <dd>{dateLabel(book.cycle.finishedOn)}</dd>
          </dl>
        </>
      }
    >
      {editions}
      {!!book.edition.catalog?.warnings?.length && (
        <ul className="catalog-warnings">
          {book.edition.catalog.warnings.map((warning, index) => (
            <li key={index}>{warning}</li>
          ))}
        </ul>
      )}
      <section className="portal-card profile-notes">
        <div className="card-heading">
          <h2>Notas personales</h2>
          <button
            className="small-save"
            disabled={busy || notes === book.notes}
            onClick={() => void onNotes(book, notes)}
          >
            Guardar notas
          </button>
        </div>
        <label className="inline-notes">
          <span className="sr-only">Notas personales del {labels.item}</span>
          <textarea
            value={notes}
            disabled={busy}
            rows={4}
            placeholder="Tus impresiones, ideas o recordatorios…"
            onChange={(event) => setNotes(event.target.value)}
          />
        </label>
      </section>
    </BookProfile>
  );
}
