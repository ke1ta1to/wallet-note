import { Stack, Text, Title } from "@mantine/core";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/_authed/orgs/$orgId/categories")({
  component: CategoriesPage,
});

function CategoriesPage() {
  return (
    <Stack>
      <Title order={2}>カテゴリ</Title>
      <Text c="dimmed">(未実装)</Text>
    </Stack>
  );
}
