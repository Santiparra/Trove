import type { ReactNode } from "react";
import type {
  Book,
  Series,
} from "../../bindings/github.com/Santiparra/Trove/internal/books/models";
import type { Snapshot } from "../../bindings/github.com/Santiparra/Trove/internal/catalog/models";
import CatalogImage from "../catalog/CatalogImage";
import Icon from "../components/Icon";
import { seriesLabel } from "./presentation";
import DescriptionText from "../components/DescriptionText";
import { ExternalRating } from "./RatingEditor";
import ClassificationEditor from "../catalog/ClassificationEditor";

export default function BookProfile({
  book = null,
  snapshot = book?.edition.catalog,
  busy,
  onBack,
  onSeries,
  onSearch,
  management,
  children,
  ariaLabel,
  subtitle,
  description,
  coverURL,
  editionInfo,
  sourceDetails,
  formatLabel,
  loading = false,
  sourcesBeforeChildren = false,
  metadataActions,
  compactManagement = false,
  ratingBadge,
  subtitleDetails,
  beforeBody,
  metadataLast = false,
  classification,
}: {
  book?: Book | null;
  snapshot?: Snapshot | null;
  busy: boolean;
  onBack: () => void;
  onSeries?: (series: Series) => void;
  onSearch: (query: string) => void;
  management: ReactNode;
  children?: ReactNode;
  ariaLabel?: string;
  subtitle?: string;
  description?: string;
  coverURL?: string;
  editionInfo?: {
    publisher: string;
    publishedOn: string;
    pages: number | null;
  };
  sourceDetails?: ReactNode;
  formatLabel?: string;
  loading?: boolean;
  sourcesBeforeChildren?: boolean;
  metadataActions?: ReactNode;
  compactManagement?: boolean;
  ratingBadge?: ReactNode;
  subtitleDetails?: ReactNode;
  beforeBody?: ReactNode;
  metadataLast?: boolean;
  classification?: ReactNode;
}) {
  const title = book?.title || snapshot?.edition.title || "Libro";
  const authors = book?.authors || snapshot?.work.authors || [];
  const identifiers = book?.edition.identifiers || snapshot?.identifiers || [];
  const synopsis = description ?? snapshot?.description;
  const facts = editionInfo ?? snapshot?.edition;
  const displayFormat =
    formatLabel ||
    (
      {
        paperback: "Tapa blanda",
        hardcover: "Tapa dura",
        digital: "Digital",
        audio: "Audio",
        other: "Otro",
      } as Record<string, string>
    )[book?.edition.format || snapshot?.edition.format || ""] ||
    book?.edition.format ||
    snapshot?.edition.format;
  return (
    <section
      className={`book-page unified-book-page ${compactManagement ? "compact-management" : ""}`}
      aria-label={
        ariaLabel ||
        (book?.medium === "light-novels"
          ? "Detalle del volumen"
          : "Detalle del libro")
      }
    >
      <button className="back-link" onClick={onBack} disabled={busy}>
        <Icon name="arrow" size={17} /> Volver
      </button>
      <div className="book-profile-grid">
        <div className="profile-cover">
          <CatalogImage
            url={coverURL ?? snapshot?.coverURL ?? ""}
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
                {snapshot.work.series.map((item, index) => (
                  <span key={item.id}>
                    {index > 0 && " · "}
                    <button
                      className="series-title"
                      disabled={busy}
                      onClick={() => onSearch(item.name)}
                    >
                      {item.name}
                      {item.position != null ? ` · #${item.position}` : ""}
                    </button>
                  </span>
                ))}
              </p>
            )
          )}
          <div className="profile-heading-row">
            <div className="profile-heading-copy">
              <h1>{title}</h1>
              {subtitle && <p className="novel-subtitle">{subtitle}</p>}
              {subtitleDetails}
              <p className="hero-author">
                {authors.length
                  ? authors.map((author, index) => (
                      <span key={`${author}:${index}`}>
                        {index > 0 && " · "}
                        <button
                          className="author-link"
                          disabled={busy}
                          onClick={() => onSearch(author)}
                        >
                          {author}
                        </button>
                      </span>
                    ))
                  : "Autoría sin especificar"}
              </p>
              <dl className="profile-facts">
                {(book?.edition.language || snapshot?.edition.language) && (
                  <>
                    <dt>Idioma</dt>
                    <dd>
                      {book?.edition.language || snapshot?.edition.language}
                    </dd>
                  </>
                )}
                {(book?.edition.format || snapshot?.edition.format) && (
                  <>
                    <dt>Formato</dt>
                    <dd>{displayFormat}</dd>
                  </>
                )}
                {facts?.publisher && (
                  <>
                    <dt>Editorial</dt>
                    <dd>{facts.publisher}</dd>
                  </>
                )}
                {facts?.pages != null && (
                  <>
                    <dt>Páginas</dt>
                    <dd>{facts.pages}</dd>
                  </>
                )}
                {facts?.publishedOn && (
                  <>
                    <dt>Publicación</dt>
                    <dd>{facts.publishedOn}</dd>
                  </>
                )}
              </dl>
              {loading && (
                <div
                  className="profile-loading-skeleton"
                  role="status"
                  aria-label="Cargando información del volumen"
                  aria-busy="true"
                >
                  <span />
                  <span />
                  <span />
                  <span />
                  <span />
                </div>
              )}
              {synopsis && <DescriptionText text={synopsis} />}
            </div>
            {ratingBadge === undefined
              ? snapshot && (
                  <ExternalRating
                    rating={snapshot.work.rating}
                    provider="Hardcover · obra"
                  />
                )
              : ratingBadge}
          </div>
        </div>
        <aside className="profile-management" aria-label="Gestión del libro">
          {management}
        </aside>
      </div>
      {classification === undefined && snapshot?.work.classification ? (
        <ClassificationEditor
          title="Géneros y etiquetas"
          effective={snapshot.work.classification}
        />
      ) : (
        classification
      )}
      {beforeBody}
      {!metadataLast && metadataActions}
      {(metadataLast || !sourcesBeforeChildren) && children}
      {metadataLast && metadataActions}
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
        {sourceDetails}
      </details>
      {!metadataLast && sourcesBeforeChildren && children}
    </section>
  );
}
