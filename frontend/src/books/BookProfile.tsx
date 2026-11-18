import { useEffect, useState, type ReactNode } from "react";
import type {
  Book,
  Series,
} from "../../bindings/github.com/Santiparra/Trove/internal/books/models";
import type { Snapshot } from "../../bindings/github.com/Santiparra/Trove/internal/catalog/models";
import CatalogImage from "../catalog/CatalogImage";
import Icon from "../components/Icon";
import { seriesLabel } from "./presentation";

export default function BookProfile({
  book = null,
  snapshot = book?.edition.catalog,
  busy,
  onBack,
  onSeries,
  management,
  children,
}: {
  book?: Book | null;
  snapshot?: Snapshot | null;
  busy: boolean;
  onBack: () => void;
  onSeries?: (series: Series) => void;
  management: ReactNode;
  children?: ReactNode;
}) {
  const [expanded, setExpanded] = useState(false);
  useEffect(() => setExpanded(false), [book?.id, snapshot?.edition.id]);
  const title = book?.title || snapshot?.edition.title || "Libro";
  const authors = book?.authors || snapshot?.work.authors || [];
  const identifiers = book?.edition.identifiers || snapshot?.identifiers || [];
  return (
    <section
      className="book-page unified-book-page"
      aria-label={
        book?.medium === "light-novels"
          ? "Detalle del volumen"
          : "Detalle del libro"
      }
    >
      <button className="back-link" onClick={onBack} disabled={busy}>
        <Icon name="arrow" size={17} /> Volver
      </button>
      <div className="book-profile-grid">
        <div className="profile-cover">
          <CatalogImage
            url={snapshot?.coverURL || ""}
            fallback={
              <div className="cover-placeholder" aria-hidden="true">
                <Icon name="books" size={44} />
                <span className="cover-title">{title}</span>
                <span className="cover-caption">SIN PORTADA</span>
              </div>
            }
          />
        </div>
        <div className="profile-info">
          {book?.series ? (
            <button
              className="series-title"
              disabled={busy}
              onClick={() => onSeries?.(book.series!)}
            >
              {seriesLabel(book)}
            </button>
          ) : (
            !!snapshot?.work.series?.length && (
              <p className="profile-series">
                {snapshot.work.series
                  .map(
                    (item) =>
                      `${item.name}${item.position != null ? ` · #${item.position}` : ""}`,
                  )
                  .join(" · ")}
              </p>
            )
          )}
          <h1>{title}</h1>
          <p className="hero-author">
            {authors.join(" · ") || "Autoría sin especificar"}
          </p>
          <dl className="profile-facts">
            {(book?.edition.language || snapshot?.edition.language) && (
              <>
                <dt>Idioma</dt>
                <dd>{book?.edition.language || snapshot?.edition.language}</dd>
              </>
            )}
            {(book?.edition.format || snapshot?.edition.format) && (
              <>
                <dt>Formato</dt>
                <dd>
                  {(
                    {
                      paperback: "Tapa blanda",
                      hardcover: "Tapa dura",
                      digital: "Digital",
                      audio: "Audio",
                      other: "Otro",
                    } as Record<string, string>
                  )[book?.edition.format || snapshot?.edition.format || ""] ||
                    book?.edition.format ||
                    snapshot?.edition.format}
                </dd>
              </>
            )}
            {snapshot?.edition.publisher && (
              <>
                <dt>Editorial</dt>
                <dd>{snapshot.edition.publisher}</dd>
              </>
            )}
            {snapshot?.edition.pages != null && (
              <>
                <dt>Páginas</dt>
                <dd>{snapshot.edition.pages}</dd>
              </>
            )}
            {snapshot?.edition.publishedOn && (
              <>
                <dt>Publicación</dt>
                <dd>{snapshot.edition.publishedOn}</dd>
              </>
            )}
          </dl>
          {snapshot?.description && (
            <div className="profile-description">
              <p
                className={
                  !expanded && snapshot.description.length > 600
                    ? "description-collapsed"
                    : ""
                }
              >
                {snapshot.description}
              </p>
              {snapshot.description.length > 600 && (
                <button
                  className="text-action"
                  aria-expanded={expanded}
                  onClick={() => setExpanded((value) => !value)}
                >
                  {expanded ? "Leer menos" : "Leer más"}
                </button>
              )}
            </div>
          )}
        </div>
        <aside className="profile-management" aria-label="Gestión del libro">
          {management}
        </aside>
      </div>
      {children}
      <details className="profile-identifiers">
        <summary>Identificadores y fuentes</summary>
        {book?.edition.identifier && (
          <p>ISBN / ID local: {book.edition.identifier}</p>
        )}
        <div className="catalog-identifiers">
          {identifiers.map((id) => (
            <span key={`${id.scheme}:${id.value}`}>
              {id.scheme}: {id.value}
              {id.derived ? " (equivalente)" : ""}
            </span>
          ))}
        </div>
        {snapshot && (
          <>
            <p>
              {snapshot.provider} · Obra {snapshot.work.id} · Edición{" "}
              {snapshot.edition.id} ·{" "}
              {new Date(snapshot.fetchedAt).toLocaleString("es")}
            </p>
            {Object.entries(snapshot.sources || {}).map(([field, source]) => (
              <p key={field}>
                {field}: {source?.provider} · {source?.entity} {source?.id}
              </p>
            ))}
          </>
        )}
      </details>
    </section>
  );
}
