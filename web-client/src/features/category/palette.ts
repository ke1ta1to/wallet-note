// Single source of truth for category color choices.
// Mantine の各 color の shade 6 を採用 (rainbow order)。
// 詳細とユーザ自由色を受け付けない方針は docs/05_web-client.md を参照。
export const categoryPalette = [
  "#fa5252", // red
  "#fd7e14", // orange
  "#fab005", // yellow
  "#82c91e", // lime
  "#40c057", // green
  "#12b886", // teal
  "#15aabf", // cyan
  "#228be6", // blue
  "#4c6ef5", // indigo
  "#7950f2", // violet
  "#be4bdb", // grape
  "#e64980", // pink
] as const;

export type CategoryColor = (typeof categoryPalette)[number];
