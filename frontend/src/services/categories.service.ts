import { api, USE_MOCK } from "@/lib/axios";
import type { Category } from "@/types";

const delay = (ms = 300) => new Promise((r) => setTimeout(r, ms));

const mockCategories: Category[] = [
  { id: 1, slug: "development", nameEn: "Development", nameUz: "Dasturlash", nameRu: "Разработка" },
  { id: 2, slug: "design", nameEn: "Design", nameUz: "Dizayn", nameRu: "Дизайн" },
  { id: 3, slug: "business", nameEn: "Business", nameUz: "Biznes", nameRu: "Бизнес" },
  { id: 4, slug: "marketing", nameEn: "Marketing", nameUz: "Marketing", nameRu: "Маркетинг" },
  { id: 5, slug: "data-science", nameEn: "Data Science", nameUz: "Data Science", nameRu: "Наука о данных" },
  { id: 6, slug: "photography", nameEn: "Photography", nameUz: "Fotografiya", nameRu: "Фотография" },
];

export interface CreateCategoryInput {
  name_uz: string;
  name_ru: string;
  name_en: string;
  parent_id?: number | null;
}

export const categoriesService = {
  async list(): Promise<Category[]> {
    if (USE_MOCK) {
      await delay();
      return mockCategories;
    }
    const { data } = await api.get("/categories");
    return data.categories;
  },

  // Maps to the Go backend POST /v1/categories handler.
  async create(input: CreateCategoryInput): Promise<Category> {
    if (USE_MOCK) {
      await delay();
      return {
        id: Date.now(),
        slug: input.name_en.toLowerCase().replace(/[^a-z0-9]+/g, "-").replace(/^-+|-+$/g, ""),
        nameEn: input.name_en,
        nameUz: input.name_uz,
        nameRu: input.name_ru,
        parentId: input.parent_id ?? null,
      };
    }
    const { data } = await api.post("/categories", input);
    return data.category;
  },
};
