import type { Candidate } from "../../bindings/github.com/Santiparra/Trove/internal/novelcatalog/models";
import type { Novel } from "../../bindings/github.com/Santiparra/Trove/internal/novels/models";
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
