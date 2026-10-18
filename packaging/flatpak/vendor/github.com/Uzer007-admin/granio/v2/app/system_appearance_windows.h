#ifndef FYNE_SYSTEM_APPEARANCE_WINDOWS_H
#define FYNE_SYSTEM_APPEARANCE_WINDOWS_H
#include <stdint.h>

typedef struct FyneAppearance FyneAppearance;
enum {
    FYNE_APPEARANCE_COLORS = 1,
    FYNE_APPEARANCE_TEXT = 2,
    FYNE_APPEARANCE_CONTRAST = 4,
    FYNE_APPEARANCE_MOTION = 8,
    FYNE_APPEARANCE_EFFECTS = 16
};
typedef struct {
    uint32_t available, subscriptions, winrt_subscriptions, win32_subscriptions, accent, background;
    int32_t event_errors[5];
    uint32_t colors[8]; // Window, WindowText, ButtonFace/Text, Highlight/Text, GrayText, Hotlight
    double text_scale;
    int high_contrast, animations, transparency;
} FyneAppearanceSnapshot;

// Open/read/wait/close share one MTA thread. Cancel/notify may run on any thread.
FyneAppearance *fyne_appearance_open(int32_t *error);
void fyne_appearance_read(FyneAppearance *, FyneAppearanceSnapshot *);
int fyne_appearance_wait(FyneAppearance *, uint32_t timeout_ms);
void fyne_appearance_cancel(FyneAppearance *);
void fyne_appearance_close(FyneAppearance *);
void fyne_appearance_notify(FyneAppearance *);
int fyne_appearance_post_settings_message(FyneAppearance *);
void fyne_appearance_resources(long *providers, long *delegates, long *wakes);
#endif
