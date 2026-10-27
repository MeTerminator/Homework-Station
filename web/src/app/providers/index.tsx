import type { ReactNode } from "react";
import { BrowserRouter } from "react-router-dom";
import { SessionProvider } from "./session-provider";
import { ThemeProvider } from "./theme-provider";
export function AppProviders({ children }: { children: ReactNode }) {
  return (
    <BrowserRouter>
      <ThemeProvider>
        <SessionProvider>{children}</SessionProvider>
      </ThemeProvider>
    </BrowserRouter>
  );
}
