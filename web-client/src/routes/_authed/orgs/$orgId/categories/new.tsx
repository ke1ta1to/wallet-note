import { categoryPalette } from "@/features/category/palette";
import {
  NewCategorySchema,
  createCategoryMutation,
} from "@/features/category/mutations";
import { categoriesQuery } from "@/features/category/queries";
import {
  Alert,
  Button,
  ColorSwatch,
  Drawer,
  Group,
  SegmentedControl,
  Stack,
  Text,
  TextInput,
} from "@mantine/core";
import { schemaResolver, useForm } from "@mantine/form";
import { IconCheck } from "@tabler/icons-react";
import {
  useMutation,
  useQueryClient,
  useSuspenseQuery,
} from "@tanstack/react-query";
import { createFileRoute, useNavigate } from "@tanstack/react-router";

export const Route = createFileRoute("/_authed/orgs/$orgId/categories/new")({
  component: NewCategoryDrawer,
});

function NewCategoryDrawer() {
  const { orgId } = Route.useParams();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { data: existing } = useSuspenseQuery(categoriesQuery(orgId));

  const usedColors = new Set(existing.items.map((c) => c.color));
  const allUsed = categoryPalette.every((hex) => usedColors.has(hex));

  const close = () =>
    navigate({ to: "/orgs/$orgId/categories", params: { orgId } });

  const mutation = useMutation({
    ...createCategoryMutation(orgId),
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: categoriesQuery(orgId).queryKey,
      });
      close();
    },
  });

  const form = useForm({
    initialValues: {
      name: "",
      kind: "expense" as "expense" | "income",
      color:
        categoryPalette.find((hex) => !usedColors.has(hex)) ??
        categoryPalette[0],
    },
    validate: schemaResolver(NewCategorySchema, { sync: true }),
  });

  const submitError = mutation.isError
    ? ((mutation.error as { message?: string })?.message ??
      "作成に失敗しました")
    : null;

  return (
    <Drawer
      opened
      onClose={close}
      position="bottom"
      size="auto"
      title="新規カテゴリ"
    >
      <form onSubmit={form.onSubmit((values) => mutation.mutate(values))}>
        <Stack>
          <TextInput
            label="名前"
            {...form.getInputProps("name")}
            maxLength={100}
            required
            autoFocus
          />
          <SegmentedControl
            fullWidth
            {...form.getInputProps("kind")}
            data={[
              { label: "支出", value: "expense" },
              { label: "収入", value: "income" },
            ]}
          />
          <Stack gap={6}>
            <Text size="sm">色</Text>
            <Group gap="xs">
              {categoryPalette.map((hex) => {
                const isUsed = usedColors.has(hex);
                const isSelected = !isUsed && form.values.color === hex;
                return (
                  <ColorSwatch
                    key={hex}
                    component="button"
                    type="button"
                    color={hex}
                    size={36}
                    onClick={() => !isUsed && form.setFieldValue("color", hex)}
                    disabled={isUsed}
                    style={{
                      cursor: isUsed ? "not-allowed" : "pointer",
                      opacity: isUsed ? 0.25 : 1,
                      outline: isSelected
                        ? "3px solid var(--mantine-color-text)"
                        : "none",
                      outlineOffset: 2,
                    }}
                    aria-label={isUsed ? `${hex} (使用中)` : hex}
                    aria-pressed={isSelected}
                    title={isUsed ? "使用中" : undefined}
                  >
                    {isSelected && <IconCheck size={16} color="white" />}
                  </ColorSwatch>
                );
              })}
            </Group>
            {allUsed && (
              <Alert color="yellow">
                全色が使用中です。既存カテゴリを編集または削除してください。
              </Alert>
            )}
            {form.errors.color && (
              <Text size="xs" c="red">
                {form.errors.color}
              </Text>
            )}
          </Stack>
          {submitError && <Alert color="red">{submitError}</Alert>}
          <Group justify="flex-end">
            <Button variant="subtle" onClick={close} type="button">
              キャンセル
            </Button>
            <Button
              type="submit"
              loading={mutation.isPending}
              disabled={allUsed}
            >
              作成
            </Button>
          </Group>
        </Stack>
      </form>
    </Drawer>
  );
}
