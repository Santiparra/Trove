import type { Candidate } from "../../bindings/github.com/Santiparra/Trove/internal/novelcatalog/models";
import type { Novel } from "../../bindings/github.com/Santiparra/Trove/internal/novels/models";
export const novelColors: Record<string, string> = {
  "to-read": "#9caac0",
  "up-next": "#bb91ff",
  reading: "#40d98b",
  read: "#5aa9ff",
  paused: "#f4c34f",
  "did-not-finish": "#f1747e",
};
export function printedLabel(title: string, original = "") {
  const number =
    title.match(/\b(?:vol(?:ume)?\.?|tomo)\s*(\d+(?:[.,]\d+)?)/i)?.[1] ||
    original.match(/第?\s*(\d+(?:[.,]\d+)?)\s*巻/)?.[1];
  if (!number) return "";
  const part = title.match(/\b(?:part|parte)\s*(\d+(?:\.\d+)?)/i)?.[1];
  if (part) return `Parte ${part} · ${number}`;
  const year =
    title.match(/\byear\s*(\d+)/i)?.[1] ||
    title.match(/\b(\d+)(?:st|nd|rd|th)\s+year/i)?.[1];
  return year ? `Año ${year} · ${number}` : number.replace(",", ".");
}
export function volumeLabel(n: Novel, number: number) {
  const volume = n.volumes.find((v) => v.number === number);
  if (volume?.label) return volume.label;
  if (volume?.catalog) {
    const label =
      volume.catalog.label ||
      printedLabel(volume.catalog.title, volume.catalog.originalTitle);
    if (label) return label;
  }
  const ref = n.catalog?.books.find((b) => b.number === number);
  if (ref)
    return (
      ref.label ||
      printedLabel(ref.title, ref.originalTitle) ||
      `Lectura #${number}`
    );
  return String(number);
}
export function representedNovel(candidate: Candidate, novels: Novel[]) {
  return novels.some(
    (n) =>
      n.catalog?.id === candidate.id &&
      (!candidate.matchedReleaseID ||
        n.volumes.some(
          (v) =>
            v.catalog?.bookID === candidate.matchedBookID &&
            v.catalog.releaseID === candidate.matchedReleaseID,
        )),
  );
}
