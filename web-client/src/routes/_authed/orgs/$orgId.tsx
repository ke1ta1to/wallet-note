import { getOrgQuery } from "@/features/organization/queries";
import {
  ActionIcon,
  Alert,
  AppShell,
  Group,
  Menu,
  Stack,
  Text,
} from "@mantine/core";
import {
  IconCategory,
  IconChartBar,
  IconReceipt,
  IconSettings,
  IconSwitchHorizontal,
} from "@tabler/icons-react";
import { useSuspenseQuery } from "@tanstack/react-query";
import { Link, Outlet, createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/_authed/orgs/$orgId")({
  loader: ({ context: { queryClient }, params }) =>
    queryClient.ensureQueryData(getOrgQuery(params.orgId)),
  component: OrgLayout,
  errorComponent: OrgError,
});

const tabs = [
  { to: "/orgs/$orgId/transactions", icon: IconReceipt, label: "取引" },
  { to: "/orgs/$orgId/summary", icon: IconChartBar, label: "集計" },
  { to: "/orgs/$orgId/categories", icon: IconCategory, label: "カテゴリ" },
  { to: "/orgs/$orgId/settings", icon: IconSettings, label: "設定" },
] as const;

function OrgLayout() {
  const { orgId } = Route.useParams();
  const { data: org } = useSuspenseQuery(getOrgQuery(orgId));

  return (
    <AppShell header={{ height: 56 }} footer={{ height: 64 }} padding="md">
      <AppShell.Header>
        <Group h="100%" px="md" justify="space-between" wrap="nowrap">
          <Text fw={500} size="lg" truncate>
            {org.name}
          </Text>
          <Menu position="bottom-end">
            <Menu.Target>
              <ActionIcon variant="subtle" aria-label="組織を切り替え">
                <IconSwitchHorizontal size={18} />
              </ActionIcon>
            </Menu.Target>
            <Menu.Dropdown>
              <Menu.Item component={Link} to="/orgs">
                組織一覧
              </Menu.Item>
            </Menu.Dropdown>
          </Menu>
        </Group>
      </AppShell.Header>

      <AppShell.Main>
        <Outlet />
      </AppShell.Main>

      <AppShell.Footer>
        <Group h="100%" gap={0} wrap="nowrap">
          {tabs.map(({ to, icon: Icon, label }) => (
            <Link
              key={to}
              to={to}
              params={{ orgId }}
              style={{ flex: 1, textDecoration: "none", height: "100%" }}
            >
              {({ isActive }) => (
                <Stack
                  h="100%"
                  align="center"
                  justify="center"
                  gap={2}
                  c={
                    isActive ? "var(--mantine-primary-color-filled)" : "dimmed"
                  }
                >
                  <Icon size={20} />
                  <Text size="xs">{label}</Text>
                </Stack>
              )}
            </Link>
          ))}
        </Group>
      </AppShell.Footer>
    </AppShell>
  );
}

function OrgError({ error }: { error: unknown }) {
  const message =
    (error as { message?: string })?.message ?? "組織の読み込みに失敗しました";
  return (
    <AppShell padding="md">
      <AppShell.Main>
        <Alert color="red">{message}</Alert>
      </AppShell.Main>
    </AppShell>
  );
}
