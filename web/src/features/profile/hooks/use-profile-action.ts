import { perform } from "@/lib/actions";
import { useState } from "react";
export function useProfileAction() {
  const [busy, setBusy] = useState("");
  const act = (key: string, fn: () => Promise<void>, message: string) => {
    setBusy(key);
    perform(fn, message).finally(() => setBusy(""));
  };
  return { busy, act };
}
