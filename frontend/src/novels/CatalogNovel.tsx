import { useEffect, useState } from "react";
import * as Novels from "../../bindings/github.com/Santiparra/Trove/internal/desktop/novels";
import type {
  NovelPreview,
  Info,
} from "../../bindings/github.com/Santiparra/Trove/internal/desktop/models";
import type { Novel } from "../../bindings/github.com/Santiparra/Trove/internal/novels/models";
import type {
  Publication,
  Release,
} from "../../bindings/github.com/Santiparra/Trove/internal/novelcatalog/models";
import type { Book } from "../../bindings/github.com/Santiparra/Trove/internal/books/models";
import BookProfile from "../books/BookProfile";
import { errorMessage } from "../books/presentation";

export function publicationFormat(format: string) {
  return (
    (
      { print: "Impreso", digital: "Digital", audio: "Audiolibro" } as Record<
        string,
        string
      >
    )[format] || format
  );
}
export function PublicationSources({
  publication: p,
}: {
  publication: Publication;
}) {
  return (
    <>
      <p>
        Novela ligera · RanobeDB · Serie {p.seriesID} · Libro {p.bookID}
        {p.releaseID
          ? ` · Publicación ${p.releaseID}`
          : " · Sin publicación seleccionada"}
      </p>
      <p>
        Orden original: {p.order} · Consultado:{" "}
        {new Date(p.fetchedAt).toLocaleString("es")}
      </p>
      {p.releaseTitle && <p>Título de publicación: {p.releaseTitle}</p>}
      {p.hardcoverCategory && (
        <p>Categoría Hardcover verificada: {p.hardcoverCategory}</p>
      )}
      {p.isbn && <p>ISBN-13: {p.isbn}</p>}
      {Object.entries(p.sources || {}).map(
        ([field, source]) =>
          source && (
            <p key={field}>
              {field}: {source.provider} · {source.entity} {source.id}
            </p>
          ),
      )}
    </>
  );
}

export default function CatalogNovel({
  initial,
  target,
  info,
  localOnly,
  onBack,
  onDone,
  onBusy,
  onSearch,
}: {
  initial: NovelPreview;
  target: Novel | null;
  info: Info;
  localOnly: boolean;
  onBack: () => void;
  onDone: (novel: Novel, message: string, number: number) => void;
  onBusy: (value: boolean) => void;
  onSearch: (name: string) => void;
}) {
  const [preview, setPreview] = useState(initial);
  const [number, setNumber] = useState(initial.publication.number);
  const [choices, setChoices] = useState<Release[]>([]);
  const [picking, setPicking] = useState(false);
  const [choicesLoading, setChoicesLoading] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [status, setStatus] = useState("to-read");
  const [completed, setCompleted] = useState("0");
  const series = preview.series;
  const p = preview.publication;
  useEffect(() => {
    let active = true;
    setChoices([]);
    setError("");
    if (!picking || localOnly) return;
    setChoicesLoading(true);
    Novels.CatalogReleases(preview.token, number)
      .then((items) => {
        if (active) setChoices(items);
      })
      .catch((reason) => {
        if (active) setError(errorMessage(reason));
      })
      .finally(() => {
        if (active) setChoicesLoading(false);
      });
    return () => {
      active = false;
    };
  }, [picking, number, preview.token, localOnly]);
  async function action(operation: () => PromiseLike<void>) {
    if (busy) return;
    setBusy(true);
    onBusy(true);
    setError("");
    try {
      await operation();
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setBusy(false);
      onBusy(false);
    }
  }
  const local = target || preview.match.novel;
  function add(selected: string) {
    setStatus(selected);
    void action(async () => {
      const match = await Novels.AddCatalog(
        preview.token,
        selected,
        completed === "" ? null : Number(completed),
      );
      if (match.kind === "review") {
        setPreview({ ...preview, match });
      } else if (match.novel) onDone(match.novel, match.message, p.number);
    });
  }
  const book: Book = {
    id: "",
    medium: "light-novels",
    title: series.title,
    authors: series.authors.map((author) => author.name),
    series: null,
    edition: {
      id: "",
      language: p.language,
      format: p.format,
      identifier: "",
      identifiers: [],
      catalog: null,
      owned: false,
    },
    cycle: { id: "", status: "to-read", startedOn: "", finishedOn: "" },
    notes: "",
    favorite: false,
    addedOn: "",
  };
  return (
    <BookProfile
      book={book}
      ariaLabel="Detalle de novela de catálogo"
      subtitle={`Novela ligera · Volumen ${p.number} · ${p.title}`}
      description={p.description || series.description}
      coverURL={p.coverURL}
      editionInfo={p}
      formatLabel={
        p.releaseFormat === "print" && p.format === "other"
          ? "Impreso"
          : undefined
      }
      sourceDetails={<PublicationSources publication={p} />}
      busy={busy}
      onBack={onBack}
      onSearch={onSearch}
      management={
        <>
          <strong className="management-hint">
            {p.releaseID
              ? `${p.language} · ${publicationFormat(p.releaseFormat)}`
              : "Sin publicación de idioma/formato preferido"}
          </strong>
          {local ? (
            <>
              <p className="local-match">
                {target ? "Vincular con" : "Coincidencia local"}:{" "}
                <strong>{local.title}</strong>
              </p>
              {preview.match.message && (
                <p className="management-hint">{preview.match.message}</p>
              )}
              {(!preview.match.novel ||
                preview.match.novel.id === local.id ||
                preview.match.kind !== "existing") && (
                <button
                  className="primary"
                  disabled={busy}
                  onClick={() =>
                    void action(async () =>
                      onDone(
                        await Novels.LinkCatalog(local.id, preview.token),
                        "Publicación vinculada sin cambiar tu seguimiento.",
                        p.number,
                      ),
                    )
                  }
                >
                  Vincular publicación
                </button>
              )}
              {preview.match.novel && (
                <button
                  className="quiet"
                  disabled={busy}
                  onClick={() => onDone(preview.match.novel!, "", p.number)}
                >
                  Abrir registro existente
                </button>
              )}
            </>
          ) : (
            <>
              <div className="add-status">
                <button
                  className="primary"
                  disabled={busy}
                  onClick={() => add(status)}
                >
                  Agregar ·{" "}
                  {
                    info.statuses.find((option) => option.value === status)
                      ?.label
                  }
                </button>
                <select
                  aria-label="Estado al agregar novela"
                  disabled={busy}
                  value={status}
                  onChange={(event) => add(event.target.value)}
                >
                  {info.statuses.map((option) => (
                    <option value={option.value} key={option.value}>
                      {option.value === "read" ? "Leída" : option.label}
                    </option>
                  ))}
                </select>
              </div>
              <label className="catalog-series">
                Volúmenes completados
                <input
                  aria-label="Volúmenes completados al agregar"
                  type="number"
                  min={0}
                  max={10000}
                  step={1}
                  placeholder="Desconocidos"
                  value={completed}
                  disabled={busy}
                  onChange={(event) => setCompleted(event.target.value)}
                />
              </label>
              <small className="management-hint">
                El catálogo no establece tu total ni marca volúmenes leídos
                automáticamente.
              </small>
            </>
          )}
          <button
            className="secondary"
            disabled={busy || localOnly}
            onClick={() => setPicking((value) => !value)}
          >
            Elegir volumen o publicación
          </button>
          {error && (
            <p className="error-banner" role="alert">
              {error}
            </p>
          )}
        </>
      }
    >
      {picking && (
        <section
          className="edition-picker"
          aria-label="Publicaciones de novela"
        >
          <div className="card-heading">
            <h2>Volumen y publicación</h2>
            <button
              className="quiet"
              disabled={busy}
              onClick={() => setPicking(false)}
            >
              Cerrar publicaciones
            </button>
          </div>
          <label className="catalog-series">
            Orden de lectura
            <select
              aria-label="Volumen de catálogo"
              disabled={busy || localOnly}
              value={number}
              onChange={(event) => setNumber(Number(event.target.value))}
            >
              {series.books.map((ref) => (
                <option key={ref.id} value={ref.number}>
                  {ref.number} · {ref.title || ref.originalTitle}
                </option>
              ))}
            </select>
          </label>
          <p className="management-hint">
            Se conserva el orden y los títulos originales, incluidas partes. No
            se interpreta la cantidad de resultados como total traducido.
          </p>
          {choicesLoading && <p role="status">Consultando publicaciones…</p>}
          <div className="edition-grid">
            {choices.map((release) => (
              <button
                key={release.id}
                className="edition-choice"
                disabled={busy || localOnly}
                onClick={() =>
                  void action(async () => {
                    const value = await Novels.PreviewPublication(
                      preview.token,
                      number,
                      release.id,
                    );
                    setPreview(value);
                    setPicking(false);
                  })
                }
              >
                <strong>{release.title}</strong>
                <span>
                  {release.language} · {publicationFormat(release.format)} ·{" "}
                  {release.publishedOn || "Fecha desconocida"}
                </span>
                <span>
                  {release.isbn ? `ISBN ${release.isbn}` : "Sin ISBN"}
                </span>
              </button>
            ))}
          </div>
          <button
            className="quiet"
            disabled={busy || localOnly}
            onClick={() =>
              void action(async () => {
                setPreview(
                  await Novels.PreviewPublication(preview.token, number, ""),
                );
                setPicking(false);
              })
            }
          >
            Ver selección preferida del volumen
          </button>
        </section>
      )}
      {!!p.warnings?.length && (
        <ul className="catalog-warnings">
          {p.warnings.map((warning, index) => (
            <li key={index}>{warning}</li>
          ))}
        </ul>
      )}
    </BookProfile>
  );
}
