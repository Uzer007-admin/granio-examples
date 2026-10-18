# granio-tutorials

Компонент независимого проекта **Granio — Гранио**.

**Исходный репозиторий:** [tutorials](https://github.com/fyne-io/tutorials).
**Цель форка:** Адаптация учебных примеров к Granio.

История изменений, лицензии и сведения об исходных авторах сохранены.
Go-модули переведены на собственные пути `github.com/Uzer007-admin/…`;
минимальная версия Go — **1.27.1**. Ядро и общие API подключаются из
[`granio/v2`](https://github.com/Uzer007-admin/granio).

Состояние переноса и результаты проверок приведены в
[реестре экосистемы](https://github.com/Uzer007-admin/granio/blob/dev/granio/docs/ECOSYSTEM.md).
Наличие форка само по себе не означает проверку всех платформ и сценариев.

---

## Документация исходного проекта

Команды установки и ссылки в сохранённом тексте ниже относятся к исходнику.
Для Granio используются новые пути модулей, перечисленные выше.

# Fyne Tutorials

This repository provides the full source code of the Fyne tutorials.
We use the same example content for the video and text presentation.

Contributions to this repository are welcome through pull requests,
though it is best to first create a text tutorial in the 
[developer website repository](https://github.com/fyne-io/developer.fyne.io).

## Text content

You can find these tutorials on the [Fyne developers](https://developer.fyne.io)
website.

## Videos

Video content can be viewed in the [Fyne tutorial playlist](https://www.youtube.com/watch?v=S3T9l9QUa9I&list=PLjpijTpXl1_po-ld8jORR9k5NornDNKQk).
The tutorial videos are captured in 16:9 ratio and saved at 1080p resolution.
The cover images have the same resolution (1920x1080px).

Fyne applications should be running at "Normal" font size to be readable
in the output video. Other applications may need to be adjusted to a
suitable font size.

For example the following setup is used for many videos:

* Laptop display at 14" set to 3840x2160
* Linux X11 setup with system scale detected as 2.8
* Fyne user scale Normal / 1.0
* Full screen video capture

Exact setup will vary based on hardware/OS and output devices.

