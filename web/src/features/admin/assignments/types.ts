import type { Assignment } from "@/types/models";
export type AssignmentFieldChange = <K extends keyof Assignment>(
  key: K,
  value: Assignment[K],
) => void;
