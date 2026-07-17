import { api, USE_MOCK } from "@/lib/axios";

export interface UploadResult {
  url: string;
  filename: string;
}

export const uploadsService = {
  /**
   * Faylni backendga yuklaydi (POST /v1/uploads, multipart) va statik URL
   * qaytaradi. kind: "video" (dars) yoki "image" (kurs thumbnail'i).
   */
  async upload(
    file: File,
    kind: "video" | "image",
    onProgress?: (percent: number) => void,
  ): Promise<UploadResult> {
    if (USE_MOCK) {
      await new Promise((r) => setTimeout(r, 800));
      return { url: URL.createObjectURL(file), filename: file.name };
    }

    const form = new FormData();
    form.append("file", file);

    const { data } = await api.post("/uploads", form, {
      params: { kind },
      // Axios FormData uchun boundary'ni o'zi qo'yadi
      headers: { "Content-Type": undefined },
      onUploadProgress: (e) => {
        if (onProgress && e.total) onProgress(Math.round((e.loaded / e.total) * 100));
      },
    });
    return data;
  },
};
