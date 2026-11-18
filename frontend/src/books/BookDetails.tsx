import { useEffect, useState } from "react";
import type {
  Book,
  Series,
} from "../../bindings/github.com/Santiparra/Trove/internal/books/models";
import type { Info } from "../../bindings/github.com/Santiparra/Trove/internal/desktop/models";
import Icon from "../components/Icon";
import BookCover from "./BookCover";
import StatusSelect from "./StatusSelect";
import { dateLabel, labelFor, seriesLabel } from "./presentation";
import { readingLabels } from "./readingMedia";

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
}) {
  const [notes, setNotes] = useState(book.notes);
  const labels = readingLabels(book.medium);
  useEffect(() => {
    setNotes(book.notes);
  }, [book.id, book.notes]);
  useEffect(() => {
    onDirty(notes !== book.notes);
  }, [notes, book.notes, onDirty]);
  return (
    <section className="book-page" aria-label={labels.detail}>
      <button className="back-link" onClick={onBack} disabled={busy}>
        <Icon name="arrow" size={17} /> Volver
      </button>
      <div className="book-hero">
        <div className="hero-cover">
          <BookCover book={book} />
        </div>
        <div className="book-hero-text">
          <p className="eyebrow">EN TU BIBLIOTECA · REGISTRO LOCAL</p>
          {book.series && (
            <button
              className="series-title hero-series-title"
              onClick={() => onSeries(book.series!)}
              disabled={busy}
            >
              {seriesLabel(book)}
            </button>
          )}
          <h1>{book.title}</h1>
          <p className="hero-author">
            {book.authors.join(" · ") || "Autoría sin especificar"}
          </p>
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
          <div className="book-actions">
            <StatusSelect
              book={book}
              options={info.statuses}
              disabled={busy}
              onChange={onStatus}
            />
            <button className="secondary" onClick={onEdit} disabled={busy}>
              <Icon name="edit" size={16} /> {labels.edit}
            </button>
            <button
              className="quiet danger-text"
              onClick={() => onDelete(book)}
              disabled={busy}
            >
              <Icon name="trash" size={16} /> Eliminar {labels.item}
            </button>
          </div>
          <p className="local-hint">
            Al empezar se registra el inicio; al completar, el fin solo si hay
            inicio. Podés editar ambas fechas.
          </p>
        </div>
      </div>
      <div className="book-detail-grid">
        <section className="portal-card">
          <div className="card-heading">
            <h2>Tu lectura</h2>
          </div>
          <dl className="book-facts">
            <dt>Estado</dt>
            <dd>{labelFor(info.statuses, book.cycle.status)}</dd>
            <dt>Añadido</dt>
            <dd>{dateLabel(book.addedOn)}</dd>
            <dt>Inicio</dt>
            <dd>{dateLabel(book.cycle.startedOn)}</dd>
            <dt>Finalización</dt>
            <dd>{dateLabel(book.cycle.finishedOn)}</dd>
          </dl>
        </section>
        <section className="portal-card">
          <div className="card-heading">
            <h2>Obra y edición</h2>
          </div>
          <dl className="book-facts">
            <dt>Serie</dt>
            <dd>
              {book.series ? (
                <button
                  className="series-title"
                  onClick={() => onSeries(book.series!)}
                  disabled={busy}
                >
                  {seriesLabel(book)}
                </button>
              ) : (
                "Sin serie"
              )}
            </dd>
            <dt>Idioma</dt>
            <dd>{book.edition.language || "Sin especificar"}</dd>
            <dt>Formato</dt>
            <dd>{labelFor(info.formats, book.edition.format)}</dd>
            <dt>ISBN / ID</dt>
            <dd>{book.edition.identifier || "Sin especificar"}</dd>
            <dt>Propiedad</dt>
            <dd>
              {book.edition.owned ? "En propiedad" : "No marcado en propiedad"}
            </dd>
          </dl>
        </section>
        {book.medium === "books" && (
          <section
            className="portal-card full-width-card"
            aria-label="Metadata de catálogo"
          >
            <div className="card-heading">
              <h2>Catálogo externo</h2>
            </div>
            <div className="catalog-actions">
              <button className="secondary" disabled={busy} onClick={onCatalog}>
                {book.edition.catalog
                  ? "Cambiar edición vinculada"
                  : "Vincular edición de catálogo"}
              </button>
              {book.edition.catalog && (
                <>
                  <button className="quiet" disabled={busy} onClick={onRefresh}>
                    Refrescar metadata
                  </button>
                  <button
                    className="quiet danger-text"
                    disabled={busy}
                    onClick={onUnlink}
                  >
                    Desvincular catálogo
                  </button>
                </>
              )}
            </div>
            {book.edition.catalog ? (
              <>
                <p>
                  Snapshot local de {book.edition.catalog.provider} ·{" "}
                  {new Date(book.edition.catalog.fetchedAt).toLocaleString(
                    "es",
                  )}
                </p>
                <dl className="book-facts">
                  <dt>Obra / edición remota</dt>
                  <dd>
                    {book.edition.catalog.work.id} /{" "}
                    {book.edition.catalog.edition.id}
                  </dd>
                  <dt>Título de catálogo</dt>
                  <dd>{book.edition.catalog.edition.title}</dd>
                  <dt>Idioma / formato de catálogo</dt>
                  <dd>
                    {book.edition.catalog.edition.language || "Sin especificar"}{" "}
                    / {book.edition.catalog.edition.format || "Sin especificar"}
                  </dd>
                  <dt>Fuente de portada</dt>
                  <dd>
                    {book.edition.catalog.sources?.cover?.provider ||
                      "Sin portada informada"}
                  </dd>
                  <dt>Editorial</dt>
                  <dd>
                    {book.edition.catalog.edition.publisher ||
                      "Sin especificar"}
                  </dd>
                  <dt>Páginas</dt>
                  <dd>
                    {book.edition.catalog.edition.pages ?? "Sin especificar"}
                  </dd>
                  <dt>Publicación</dt>
                  <dd>
                    {book.edition.catalog.edition.publishedOn ||
                      "Sin especificar"}
                  </dd>
                </dl>
                {book.edition.catalog.description && (
                  <div className="catalog-description">
                    <h3>
                      Descripción{" "}
                      <small>
                        Fuente:{" "}
                        {book.edition.catalog.sources?.description?.provider ||
                          book.edition.catalog.provider}
                      </small>
                    </h3>
                    <p>{book.edition.catalog.description}</p>
                  </div>
                )}
                {!!book.edition.catalog.warnings?.length && (
                  <ul className="catalog-warnings">
                    {book.edition.catalog.warnings.map((warning, index) => (
                      <li key={index}>{warning}</li>
                    ))}
                  </ul>
                )}
              </>
            ) : (
              <p>
                Podés vincular una edición sin reemplazar tus datos personales.
                El registro manual sigue disponible sin conexión.
              </p>
            )}
            {!!book.edition.identifiers?.length && (
              <div className="catalog-identifiers">
                {book.edition.identifiers.map((identifier) => (
                  <span key={`${identifier.scheme}:${identifier.value}`}>
                    {identifier.scheme}: {identifier.value}
                    {identifier.derived ? " (equivalente)" : ""}
                  </span>
                ))}
              </div>
            )}
          </section>
        )}
        <section className="portal-card full-width-card">
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
              rows={5}
              placeholder="Tus impresiones, ideas o recordatorios…"
              onChange={(event) => setNotes(event.target.value)}
            />
          </label>
        </section>
      </div>
    </section>
  );
}
