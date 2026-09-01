// Raw-import declarations for the vitest/vite pipeline (?raw fixture loads).
declare module "*?raw" {
  const content: string;
  export default content;
}
