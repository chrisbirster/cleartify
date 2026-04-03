import type { ReactNode } from "react";

import { StatePanel } from "./StatePanel";

type EmptyStateProps = {
  title: string;
  description: string;
  eyebrow?: string;
  action?: ReactNode;
};

export function EmptyState({ title, description, eyebrow, action }: EmptyStateProps) {
  return <StatePanel action={action} description={description} eyebrow={eyebrow} surface="page" title={title} />;
}
