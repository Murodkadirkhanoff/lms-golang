import { api, USE_MOCK } from "@/lib/axios";
import type { Category } from "@/types";

const delay = (ms = 300) => new Promise((r) => setTimeout(r, ms));

// Categories are at most two levels deep: a parent (parentId === null) and its
// children. Courses are assigned to a child (leaf) category.
const mockCategories: Category[] = [
  // Parents
  { id: 1, slug: "development", courseCount: 1240, nameEn: "Development", nameUz: "Dasturlash", nameRu: "Разработка", parentId: null },
  { id: 2, slug: "design", courseCount: 680, nameEn: "Design", nameUz: "Dizayn", nameRu: "Дизайн", parentId: null },
  { id: 3, slug: "business", courseCount: 540, nameEn: "Business", nameUz: "Biznes", nameRu: "Бизнес", parentId: null },
  { id: 4, slug: "marketing", courseCount: 410, nameEn: "Marketing", nameUz: "Marketing", nameRu: "Маркетинг", parentId: null },
  { id: 5, slug: "data-science", courseCount: 390, nameEn: "Data Science", nameUz: "Data Science", nameRu: "Наука о данных", parentId: null },
  // Children
  { id: 7, slug: "web-development", nameEn: "Web Development", nameUz: "Veb dasturlash", nameRu: "Веб-разработка", parentId: 1 },
  { id: 8, slug: "mobile-development", nameEn: "Mobile Development", nameUz: "Mobil dasturlash", nameRu: "Мобильная разработка", parentId: 1 },
  { id: 9, slug: "ui-ux-design", nameEn: "UI/UX Design", nameUz: "UI/UX dizayn", nameRu: "UI/UX дизайн", parentId: 2 },
  { id: 10, slug: "graphic-design", nameEn: "Graphic Design", nameUz: "Grafik dizayn", nameRu: "Графический дизайн", parentId: 2 },
  { id: 11, slug: "entrepreneurship", nameEn: "Entrepreneurship", nameUz: "Tadbirkorlik", nameRu: "Предпринимательство", parentId: 3 },
  { id: 12, slug: "management", nameEn: "Management", nameUz: "Menejment", nameRu: "Менеджмент", parentId: 3 },
  { id: 13, slug: "digital-marketing", nameEn: "Digital Marketing", nameUz: "Raqamli marketing", nameRu: "Цифровой маркетинг", parentId: 4 },
  { id: 14, slug: "seo", nameEn: "SEO", nameUz: "SEO", nameRu: "SEO", parentId: 4 },
  { id: 15, slug: "machine-learning", nameEn: "Machine Learning", nameUz: "Mashinali o‘qitish", nameRu: "Машинное обучение", parentId: 5 },
  { id: 16, slug: "data-analysis", nameEn: "Data Analysis", nameUz: "Ma’lumotlar tahlili", nameRu: "Анализ данных", parentId: 5 },
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
