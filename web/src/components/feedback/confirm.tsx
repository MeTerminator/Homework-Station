import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { perform } from "@/lib/actions";
import { useState, type ReactNode } from "react";
export function Confirm({
  open,
  onClose,
  title,
  description,
  onConfirm,
  children,
}: {
  children?: ReactNode;
  open: boolean;
  onClose: () => void;
  title: string;
  description: string;
  onConfirm: () => Promise<void>;
}) {
  const [busy, setBusy] = useState(false);
  return (
    <Dialog
      open={open}
      onOpenChange={(v) => {
        if (!v && !busy) onClose();
      }}
    >
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription>{description}</DialogDescription>
        </DialogHeader>
        {children && <fieldset disabled={busy}>{children}</fieldset>}
        <DialogFooter>
          <Button variant="outline" disabled={busy} onClick={onClose}>
            取消
          </Button>
          <Button
            variant="destructive"
            disabled={busy}
            onClick={() => {
              setBusy(true);
              perform(async () => {
                await onConfirm();
                onClose();
              }).finally(() => setBusy(false));
            }}
          >
            {busy ? "处理中…" : "确认"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
