import { apiClient } from "@/api/client";
import { mutationOptions } from "@tanstack/react-query";
import * as v from "valibot";

export const NewCategorySchema = v.object({
  name: v.pipe(
    v.string(),
    v.trim(),
    v.minLength(1, "名前は必須です"),
    v.maxLength(100, "100 文字以内で入力してください"),
  ),
  kind: v.picklist(["income", "expense"], "種別を選択してください"),
  color: v.pipe(
    v.string(),
    v.regex(/^#[0-9a-fA-F]{6}$/, "色を選択してください"),
  ),
});

export type NewCategoryInput = v.InferOutput<typeof NewCategorySchema>;

export const createCategoryMutation = (orgId: string) =>
  mutationOptions({
    mutationKey: ["orgs", orgId, "categories", "create"],
    mutationFn: async (input: NewCategoryInput) => {
      const { data, error } = await apiClient.POST(
        "/orgs/{org_id}/categories",
        {
          params: { path: { org_id: orgId } },
          body: input,
        },
      );
      if (error) throw error;
      return data;
    },
  });
