import type { Book } from "../../bindings/github.com/Santiparra/Trove/internal/books/models";
import Icon from "../components/Icon";
import CatalogImage from "../catalog/CatalogImage";

export default function BookCover({
  book,
  compact = false,
}: {
  book: Book;
  compact?: boolean;
}) {
  const fallback = (
    <div
      className={`cover-placeholder ${compact ? "compact-cover" : ""}`}
      aria-hidden="true"
    >
      <span className="cover-rule" />
      <Icon name="books" size={compact ? 22 : 30} />
      <span className="cover-title">{book.title}</span>
      <span className="cover-caption">SIN PORTADA</span>
    </div>
  );
  return (
    <CatalogImage
      url={book.edition.catalog?.coverURL || ""}
      fallback={fallback}
    />
  );
}
