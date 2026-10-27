import type { ReactNode } from "react";
import { createContext, useContext, useEffect, useState } from "react";

export type Theme = "system" | "light" | "dark";

export const ThemeContext = createContext<{
  theme: Theme;
  setTheme: (v: Theme) => void;
  dark: boolean;
}>(null!);

export const useTheme = () => useContext(ThemeContext);

export function ThemeProvider({ children }: { children: ReactNode }) {
  const [theme, setTheme] = useState<Theme>(() => {
    const t = localStorage.getItem("theme");
    return t === "light" || t === "dark" ? t : "system";
  });
  const [system, setSystem] = useState(
    matchMedia("(prefers-color-scheme: dark)").matches,
  );
  useEffect(() => {
    const q = matchMedia("(prefers-color-scheme: dark)");
    const f = () => setSystem(q.matches);
    q.addEventListener("change", f);
    return () => q.removeEventListener("change", f);
  }, []);
  const dark = theme === "dark" || (theme === "system" && system);
  useEffect(() => {
    document.documentElement.classList.toggle("dark", dark);
    localStorage.setItem("theme", theme);
  }, [theme, dark]);
  return (
    <ThemeContext.Provider value={{ theme, setTheme, dark }}>
      {children}
    </ThemeContext.Provider>
  );
}
