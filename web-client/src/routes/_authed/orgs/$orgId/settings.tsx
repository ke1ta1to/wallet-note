import { Stack, Text, Title } from "@mantine/core";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/_authed/orgs/$orgId/settings")({
  component: SettingsPage,
});

function SettingsPage() {
  return (
    <Stack>
      <Title order={2}>設定</Title>
      <Text c="dimmed">(未実装)</Text>
    </Stack>
  );
}
