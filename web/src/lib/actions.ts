import { toast } from "sonner";

export async function perform(fn: () => Promise<void>, message?: string) {
  try {
    await fn();
    if (message) toast.success(message);
  } catch (e) {
    toast.error(e instanceof Error ? e.message : "操作失败");
  }
}
