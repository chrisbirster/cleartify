import { StatePanel } from "./StatePanel";

type EmptyStateProps = {
  title: string;
  description: string;
  eyebrow?: string;
};

export function EmptyState({ title, description, eyebrow }: EmptyStateProps) {
  return <StatePanel description={description} eyebrow={eyebrow} surface="page" title={title} />;
}
