import type { Rating } from "../../bindings/github.com/Santiparra/Trove/internal/catalog/models";
import { dateLabel } from "./presentation";
import "./rating.css";

const numberLabel = (value: number) =>
  new Intl.NumberFormat("es", { maximumFractionDigits: 2 }).format(value);
export const ratingLabel = (value: number | null | undefined) =>
  value == null ? "Sin puntaje" : `${numberLabel(value)}/10`;

export function ExternalRating({
  rating,
  provider,
  compact = false,
  caption,
  secondary = false,
}: {
  rating?: Rating | null;
  provider: string;
  compact?: boolean;
  caption?: string;
  secondary?: boolean;
}) {
  const score = rating ? numberLabel((rating.value * 10) / rating.scale) : "—";
  const votes = rating?.votes;
  const quantity =
    votes == null
      ? "— valoraciones"
      : `${new Intl.NumberFormat("es").format(votes)} ${votes === 1 ? "valoración" : "valoraciones"}`;
  const detail = rating
    ? `${provider} · ${ratingLabel((rating.value * 10) / rating.scale)}\nOriginal: ${rating.value}/${rating.scale}\nConsultado ${dateLabel(rating.fetchedAt.slice(0, 10))}`
    : `${provider} · Sin puntaje disponible`;
  return (
    <div
      className={`external-rating ${compact ? "is-inline" : "is-badge"} ${secondary ? "is-secondary" : ""}`}
      aria-label={`Puntaje externo de ${provider}: ${score}/10, ${quantity}`}
      title={detail}
    >
      {caption && <small className="rating-caption">{caption}</small>}
      <strong>{score}</strong>
      <small>{quantity}</small>
    </div>
  );
}

// The selected option is persisted immediately; the parent owns the saved value.
export default function RatingEditor({
  value,
  mode,
  allowAuto = false,
  title = "Puntaje personal",
  label,
  busy,
  onSave,
}: {
  value: number | null | undefined;
  mode?: string;
  allowAuto?: boolean;
  title?: string;
  label?: string;
  busy: boolean;
  onSave: (mode: string, value: number | null) => Promise<boolean>;
}) {
  const selected =
    allowAuto && mode === "auto"
      ? "auto"
      : value == null || (allowAuto && !mode)
        ? ""
        : String(Math.floor(value));
  const fraction =
    selected !== "" &&
    selected !== "auto" &&
    value != null &&
    Number(selected) < 10
      ? Math.round((value - Math.floor(value)) * 4) / 4
      : 0;
  const descriptions = [
    "Pésimo",
    "Horrible",
    "Muy malo",
    "Malo",
    "Regular",
    "Aceptable",
    "Bueno",
    "Muy bueno",
    "Excelente",
    "Obra maestra",
  ];
  return (
    <div className="rating-select" title={title}>
      {label && <span className="rating-select-label">{label}</span>}
      <span className="rating-star" aria-hidden="true">
        ★
      </span>
      <select
        aria-label={title}
        disabled={busy}
        aria-busy={busy}
        value={selected}
        onChange={(event) => {
          const next = event.target.value;
          void onSave(
            next === "auto" ? "auto" : next === "" ? "" : "manual",
            next === "" || next === "auto"
              ? null
              : Math.min(10, Number(next) + fraction),
          );
        }}
      >
        <option value="">Sin puntuar</option>
        {allowAuto && (
          <option value="auto">
            Auto
            {mode === "auto" && value != null ? ` · ${numberLabel(value)}` : ""}
          </option>
        )}
        {Array.from({ length: 10 }, (_, i) => 10 - i).map((score) => (
          <option key={score} value={String(score)}>
            ({score}) {descriptions[score - 1]}
          </option>
        ))}
      </select>
      <select
        className="rating-decimal"
        aria-label={`Decimales de ${title.toLowerCase()}`}
        disabled={
          busy || selected === "" || selected === "auto" || selected === "10"
        }
        value={String(fraction)}
        onChange={(event) =>
          void onSave("manual", Number(selected) + Number(event.target.value))
        }
      >
        {[0, 0.25, 0.5, 0.75].map((part) => (
          <option key={part} value={String(part)}>
            ,{String(Math.round(part * 100)).padStart(2, "0")}
          </option>
        ))}
      </select>
    </div>
  );
}
