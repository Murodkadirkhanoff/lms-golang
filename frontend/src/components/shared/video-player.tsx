"use client";

import { VideoOff } from "lucide-react";
import { useT } from "@/providers/locale-provider";

interface VideoPlayerProps {
  src?: string;
  title?: string;
  poster?: string;
}

// Native HTML5 player for video lessons. Falls back to a placeholder when the
// lesson has no source yet (e.g. upload still processing).
export function VideoPlayer({ src, title, poster }: VideoPlayerProps) {
  const t = useT();

  if (!src) {
    return (
      <div className="grid aspect-video w-full place-items-center bg-black text-white">
        <div className="text-center">
          <VideoOff className="mx-auto size-12 opacity-70" />
          <p className="mt-2 text-sm text-slate-300">{t("learn.videoUnavailable")}</p>
        </div>
      </div>
    );
  }

  return (
    <video
      key={src}
      controls
      controlsList="nodownload"
      poster={poster}
      preload="metadata"
      aria-label={title}
      className="aspect-video w-full bg-black"
    >
      <source src={src} />
      {t("learn.videoUnsupported")}
    </video>
  );
}
