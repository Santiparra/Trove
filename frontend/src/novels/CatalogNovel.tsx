import { useEffect, useRef, useState } from "react";
import * as Novels from "../../bindings/github.com/Santiparra/Trove/internal/desktop/novels";
import type {
  NovelPreview,
  Info,
} from "../../bindings/github.com/Santiparra/Trove/internal/desktop/models";
import { Novel } from "../../bindings/github.com/Santiparra/Trove/internal/novels/models";
import {
  Publication,
  type Release,
} from "../../bindings/github.com/Santiparra/Trove/internal/novelcatalog/models";
import type { Book } from "../../bindings/github.com/Santiparra/Trove/internal/books/models";
import BookProfile from "../books/BookProfile";
import { errorMessage } from "../books/presentation";
import { printedLabel } from "./presentation";
import VolumeGallery from "./VolumeGallery";

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
  loadingSeries = false,
  browseError = "",
}: {
  initial: NovelPreview;
  target: Novel | null;
  info: Info;
  localOnly: boolean;
  onBack: () => void;
  onDone: (novel: Novel, message: string, number: number) => void;
  onBusy: (value: boolean) => void;
  onSearch: (name: string) => void;
  loadingSeries?: boolean;
  browseError?: string;
}) {
  const [preview, setPreview] = useState(initial);
  const [number, setNumber] = useState(initial.publication.number || 1);
  const [choices, setChoices] = useState<Release[]>([]);
  const [picking, setPicking] = useState(false);
  const [choicesLoading, setChoicesLoading] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [metadataLoading, setMetadataLoading] = useState(false);
  const [status, setStatus] = useState("to-read");
  const [completed, setCompleted] = useState("0");
  const cache = useRef(new Map<string, { value: NovelPreview; at: number }>());
  const request = useRef(0);
  const series = preview.series;
  const ready =
    !!preview.publication.bookID && preview.publication.number === number;
  const ref = series.books.find((book) => book.number === number);
  const p = ready
    ? preview.publication
    : new Publication({
        number,
        title: ref?.title || "",
        originalTitle: ref?.originalTitle || "",
        label: ref?.label || "",
        coverURL: ref?.coverURL || series.coverURL,
      });
  const local = target || preview.match.novel;
  useEffect(() => {
    if (initial.publication.bookID) {
      const item = { value: initial, at: Date.now() };
      cache.current.set(`${initial.publication.number}:`, item);
      cache.current.set(
        `${initial.publication.number}:${initial.publication.releaseID}`,
        item,
      );
    }
  }, [initial]);
  async function load(selected: number, releaseID: string) {
    if (localOnly || loadingSeries || !preview.token) return;
    const key = `${selected}:${releaseID}`;
    const saved = cache.current.get(key);
    const generation = ++request.current;
    setError("");
    setNumber(selected);
    if (saved && Date.now() - saved.at < 120000) {
      setPreview(saved.value);
      setMetadataLoading(false);
      return;
    }
    setMetadataLoading(true);
    try {
      const value = await Novels.PreviewPublication(
        preview.token,
        selected,
        releaseID,
      );
      if (generation !== request.current) return;
      cache.current.set(key, { value, at: Date.now() });
      cache.current.set(`${selected}:${value.publication.releaseID}`, {
        value,
        at: Date.now(),
      });
      setPreview(value);
    } catch (reason) {
      if (generation === request.current) setError(errorMessage(reason));
    } finally {
      if (generation === request.current) setMetadataLoading(false);
    }
  }
  useEffect(() => {
    if (!loadingSeries && preview.token && !initial.publication.bookID)
      void load(number, initial.publication.releaseID || "");
    return () => {
      request.current++;
    };
  }, [initial.token, loadingSeries, localOnly]);
  useEffect(() => {
    let active = true;
    setChoices([]);
    setChoicesLoading(false);
    if (!picking || localOnly || loadingSeries || !preview.token) return;
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
  }, [picking, number, localOnly, loadingSeries]);
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
  function add(selected: string) {
    if (!ready) return;
    setStatus(selected);
    void action(async () => {
      const match = await Novels.AddCatalog(
        preview.token,
        selected,
        completed === "" ? null : Number(completed),
      );
      if (match.kind === "review") setPreview({ ...preview, match });
      else if (match.novel)
        onDone(
          match.novel,
          match.message,
          target ? number : (match.novel.completed || 0) > 0 ? 0 : number,
        );
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
      language: ready ? p.language : "",
      format: ready ? p.format : "",
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
  const gallery = new Novel({
    id: local?.id || "",
    title: series.title,
    catalog: series,
    volumes: local?.volumes || [],
    completed: local?.completed ?? 0,
    status: local?.status || "to-read",
  });
  const label =
    p.label || printedLabel(p.title, p.originalTitle) || `Lectura #${number}`;
  return (
    <BookProfile
      book={book}
      ariaLabel="Detalle de novela de catálogo"
      subtitle={`${loadingSeries ? "Consultando novela" : "Novela ligera"} · Volumen ${label}${p.title ? ` · ${p.title}` : ""}`}
      description={ready ? p.description || series.description : undefined}
      coverURL={p.coverURL}
      editionInfo={ready ? p : undefined}
      formatLabel={
        p.releaseFormat === "print" && p.format === "other"
          ? "Impreso"
          : undefined
      }
      sourceDetails={ready ? <PublicationSources publication={p} /> : undefined}
      loading={
        loadingSeries || metadataLoading || (!ready && !error && !browseError)
      }
      busy={busy}
      onBack={onBack}
      onSearch={onSearch}
      management={
        <>
          <section
            className="management-group"
            aria-label="Gestión de biblioteca"
          >
            <h3>Tu biblioteca</h3>
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
                    disabled={
                      busy || !ready || metadataLoading || loadingSeries
                    }
                    onClick={() =>
                      void action(async () =>
                        onDone(
                          await Novels.LinkCatalog(local.id, preview.token),
                          "Publicación vinculada sin cambiar tu seguimiento.",
                          number,
                        ),
                      )
                    }
                  >
                    Vincular publicación
                  </button>
                )}
                {preview.match.novel && (
                  <button
                    className="secondary"
                    disabled={busy}
                    onClick={() => onDone(preview.match.novel!, "", number)}
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
                    disabled={
                      busy || !ready || metadataLoading || loadingSeries
                    }
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
                    disabled={
                      busy || !ready || metadataLoading || loadingSeries
                    }
                    value={status}
                    onChange={(event) => add(event.target.value)}
                  >
                    {info.statuses.map((option) => (
                      <option key={option.value} value={option.value}>
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
                    disabled={busy || loadingSeries}
                    onChange={(event) => setCompleted(event.target.value)}
                  />
                </label>
                <small className="management-hint">
                  El catálogo no establece tu total ni marca volúmenes leídos
                  automáticamente.
                </small>
              </>
            )}
          </section>
          <section
            className="management-group"
            aria-label="Publicación del volumen"
          >
            <h3>Volumen y publicación</h3>
            <strong className="management-hint">
              {ready
                ? p.releaseID
                  ? `${p.language} · ${publicationFormat(p.releaseFormat)}`
                  : "Sin publicación preferida verificada"
                : "Consultando publicación…"}
            </strong>
            <button
              className="secondary"
              disabled={busy || localOnly || loadingSeries || !preview.token}
              onClick={() => setPicking((value) => !value)}
            >
              Elegir volumen o publicación
            </button>
          </section>
          {(error || browseError) && (
            <p className="error-banner" role="alert">
              {browseError || error}
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
              value={number}
              disabled={busy || localOnly}
              onChange={(event) => void load(Number(event.target.value), "")}
            >
              {series.books.map((book) => (
                <option key={book.id} value={book.number}>
                  {book.label ||
                    printedLabel(book.title, book.originalTitle) ||
                    `Lectura #${book.number}`}{" "}
                  · {book.title || book.originalTitle}
                </option>
              ))}
            </select>
          </label>
          {choicesLoading && <p role="status">Consultando publicaciones…</p>}
          <div className="edition-grid">
            {choices.map((release) => (
              <button
                className="edition-choice"
                key={release.id}
                disabled={busy || localOnly}
                onClick={() => {
                  void load(number, release.id);
                  setPicking(false);
                }}
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
        </section>
      )}
      <section className="novel-volumes" aria-label="Volúmenes de la novela">
        <div className="card-heading">
          <h2>Volúmenes</h2>
          <span className="management-hint">
            {series.books.length
              ? `${series.books.length} disponibles en catálogo`
              : "Consultando volúmenes…"}
          </span>
        </div>
        {loadingSeries ? (
          <div
            className="gallery-loading-skeleton"
            role="status"
            aria-label="Cargando volúmenes"
          >
            <span />
            <span />
            <span />
            <span />
            <span />
          </div>
        ) : (
          <VolumeGallery
            novel={gallery}
            busy={busy || localOnly}
            next={local?.completed != null ? local.completed + 1 : number}
            onVolume={(selected) => {
              setPicking(false);
              void load(selected, "");
            }}
          />
        )}
      </section>
      {ready && !!p.warnings.length && (
        <ul className="catalog-warnings">
          {p.warnings.map((warning, index) => (
            <li key={index}>{warning}</li>
          ))}
        </ul>
      )}
    </BookProfile>
  );
}
