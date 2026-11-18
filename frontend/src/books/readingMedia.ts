// Presentation labels only. Go validates and enforces the medium for every operation.
const books = {
  medium: "books",
  title: "Libros",
  item: "libro",
  itemName: "Libro",
  plural: "libros",
  library: "Biblioteca de libros",
  detail: "Detalle del libro",
  add: "Añadir libro",
  edit: "Editar libro",
};
const novels = {
  medium: "light-novels",
  title: "Novelas ligeras",
  item: "volumen",
  itemName: "Volumen",
  plural: "volúmenes",
  library: "Biblioteca de novelas ligeras",
  detail: "Detalle del volumen",
  add: "Añadir volumen",
  edit: "Editar volumen",
};
export function readingSection(key: string) {
  return key === "books" ? books : key === "novels" ? novels : null;
}
export function readingLabels(medium: string) {
  return medium === novels.medium ? novels : books;
}
