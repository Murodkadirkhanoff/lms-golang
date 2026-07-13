"use client";

import { createContext, useCallback, useContext, useEffect, useMemo, useState } from "react";

// Bumped from the old course-only `number[]` format. Older carts are ignored.
const STORAGE_KEY = "learnhub.cart.v2";

// A cart line is either a whole course or a single lesson. Both carry the
// parent courseId so we can resolve titles/prices and link back to the course.
// Mirrors `order_items` (exactly one of course_id / lesson_id is set).
export type CartEntry =
  | { kind: "course"; courseId: number }
  | { kind: "lesson"; courseId: number; lessonId: number };

export function entryKey(entry: CartEntry): string {
  return entry.kind === "course" ? `course-${entry.courseId}` : `lesson-${entry.lessonId}`;
}

interface CartContextValue {
  items: CartEntry[];
  count: number;
  hasCourse: (courseId: number) => boolean;
  hasLesson: (lessonId: number) => boolean;
  addCourse: (courseId: number) => void;
  addLesson: (courseId: number, lessonId: number) => void;
  removeCourse: (courseId: number) => void;
  removeLesson: (lessonId: number) => void;
  clear: () => void;
}

const CartContext = createContext<CartContextValue | null>(null);

export function CartProvider({ children }: { children: React.ReactNode }) {
  const [items, setItems] = useState<CartEntry[]>([]);

  // Hydrate from localStorage on mount (client-only).
  useEffect(() => {
    try {
      const raw = localStorage.getItem(STORAGE_KEY);
      if (raw) setItems(JSON.parse(raw) as CartEntry[]);
    } catch {
      // ignore malformed storage
    }
  }, []);

  useEffect(() => {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(items));
  }, [items]);

  const hasCourse = useCallback(
    (courseId: number) => items.some((i) => i.kind === "course" && i.courseId === courseId),
    [items],
  );
  const hasLesson = useCallback(
    (lessonId: number) => items.some((i) => i.kind === "lesson" && i.lessonId === lessonId),
    [items],
  );

  const addCourse = useCallback((courseId: number) => {
    // Buying the whole course supersedes any of its individual lessons already
    // in the cart, so drop those to avoid charging twice.
    setItems((prev) => {
      const withoutLessons = prev.filter((i) => !(i.kind === "lesson" && i.courseId === courseId));
      if (withoutLessons.some((i) => i.kind === "course" && i.courseId === courseId)) {
        return withoutLessons;
      }
      return [...withoutLessons, { kind: "course", courseId }];
    });
  }, []);

  const addLesson = useCallback((courseId: number, lessonId: number) => {
    setItems((prev) => {
      // If the whole course is already in the cart, the lesson is included.
      if (prev.some((i) => i.kind === "course" && i.courseId === courseId)) return prev;
      if (prev.some((i) => i.kind === "lesson" && i.lessonId === lessonId)) return prev;
      return [...prev, { kind: "lesson", courseId, lessonId }];
    });
  }, []);

  const removeCourse = useCallback((courseId: number) => {
    setItems((prev) => prev.filter((i) => !(i.kind === "course" && i.courseId === courseId)));
  }, []);

  const removeLesson = useCallback((lessonId: number) => {
    setItems((prev) => prev.filter((i) => !(i.kind === "lesson" && i.lessonId === lessonId)));
  }, []);

  const clear = useCallback(() => setItems([]), []);

  const value = useMemo<CartContextValue>(
    () => ({
      items,
      count: items.length,
      hasCourse,
      hasLesson,
      addCourse,
      addLesson,
      removeCourse,
      removeLesson,
      clear,
    }),
    [items, hasCourse, hasLesson, addCourse, addLesson, removeCourse, removeLesson, clear],
  );

  return <CartContext.Provider value={value}>{children}</CartContext.Provider>;
}

export function useCart() {
  const ctx = useContext(CartContext);
  if (!ctx) throw new Error("useCart must be used within CartProvider");
  return ctx;
}
