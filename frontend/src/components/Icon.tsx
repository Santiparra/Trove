const paths = {
  books: [
    "M4 3h7a3 3 0 0 1 3 3v15a4 4 0 0 0-4-2H4z",
    "M14 6a3 3 0 0 1 3-3h3v16h-2a4 4 0 0 0-4 2",
  ],
  novels: ["M4 4h10v16H4z", "m12 12 7-8 2 2-7 8-3 1z", "M7 16h3"],
  anime: [
    "M6 3h12a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2Z",
    "m10 8 6 4-6 4z",
  ],
  manga: ["M4 3h16v18H4z", "M4 11h16M12 3v8M9 11v10", "m15 15 2 3"],
  manhwa: ["M5 3h14v18H5z", "M5 9h14M5 15h14", "m9 6 3 0m-3 6 6 0m-6 6 3 0"],
  tv: ["M4 7h16v13H4z", "m8 3 4 4 4-4", "M9 23h6"],
  movies: ["M3 8h18v12H3z", "M3 3h18v5H3z", "m7 3 3 5m4-5 3 5"],
  plus: ["M12 5v14M5 12h14"],
  arrow: ["M19 12H5m6-6-6 6 6 6"],
  trash: ["M3 6h18M9 6V3h6v3M5 6l1 15h12l1-15M10 10v7m4-7v7"],
  search: ["M16 10a6 6 0 1 1-12 0 6 6 0 0 1 12 0Z", "m15 15 6 6"],
  refresh: [
    "M20 7v5h-5",
    "M4 17v-5h5",
    "M5 8a8 8 0 0 1 13-3l2 3M4 16l2 3a8 8 0 0 0 13-3",
  ],
  edit: ["m4 16 12-12 4 4L8 20H4z", "m13 7 4 4"],
  close: ["m6 6 12 12M6 18 18 6"],
  chart: ["M4 4v16h17M8 16v-4m5 4V7m5 9v-6"],
  chevron: ["m9 5 7 7-7 7"],
  list: ["M9 6h11M9 12h11M9 18h11M4 6h.01M4 12h.01M4 18h.01"],
} as const;

export type IconName = keyof typeof paths;

export default function Icon({
  name,
  size = 20,
}: {
  name: IconName;
  size?: number;
}) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.6"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      {paths[name].map((d, i) => (
        <path key={i} d={d} />
      ))}
    </svg>
  );
}
