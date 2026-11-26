import type { Novel } from "../../bindings/github.com/Santiparra/Trove/internal/novels/models";
import { novelColors, volumeLabel } from "./presentation";
import { NovelCover } from "./NovelPortal";

export default function VolumeGallery({
  novel: n,
  busy,
  next,
  onVolume,
}: {
  novel: Novel;
  busy: boolean;
  next: number;
  onVolume: (number: number) => void;
}) {
  const numbers = [
    ...new Set([
      ...n.volumes.map((v) => v.number),
      ...(n.catalog?.books.map((b) => b.number) || []),
    ]),
  ].sort((a, b) => a - b);
  return (
    <div className="volume-cover-gallery" aria-label="Volúmenes disponibles">
      {numbers.map((number) => {
        const v = n.volumes.find((v) => v.number === number);
        const ref = n.catalog?.books.find((b) => b.number === number);
        const label = volumeLabel(n, number);
        const title =
          v?.title || v?.catalog?.title || ref?.title || `Volumen ${label}`;
        const state = v?.cycle.status || "unregistered";
        const run = n.readings?.find((r) => r.id === n.activeRunID);
        const pending =
          number === next &&
          (state !== "read" || (!!run && v?.cycle.runID !== run.id));
        return (
          <button
            key={ref?.id || v?.id || number}
            className={`volume-cover-item ${pending ? "is-current" : ""}`}
            disabled={busy}
            aria-label={`Abrir volumen ${label}`}
            title={`${title} · ${v ? state : "No registrado"}`}
            onClick={() => onVolume(number)}
          >
            <span className="volume-cover-image">
              <NovelCover
                url={v?.catalog?.coverURL || ref?.coverURL || ""}
                label={label}
                lazy
              />
              <span
                className={`volume-state-dot ${v ? "" : "unregistered"}`}
                style={{ background: v ? novelColors[state] : undefined }}
                aria-label={v ? `Estado: ${state}` : "No registrado"}
              />
            </span>
            <span className="volume-cover-label">{label}</span>
          </button>
        );
      })}
    </div>
  );
}
