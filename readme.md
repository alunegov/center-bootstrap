# center_bootstrap

Утилита для автоматического накатывания (Центровки)[http://rosgit:3000/ROS/Center-Droid.git] на телефон.

Процесс накатывания включает следующие шаги:
- получение свойств телефона через `adb shell getprop`
- определение пары <имя свойства с серийным номером, серийный номер>
- определение номера Центровки (последний номер плюс 1)
- определение ANDROID_ID через [CenterId-Droid](http://rosgit:3000/ROS/CenterId-Droid.git) (опционально)
- сохранение базы Центровок (номер, серийник, ANDROID_ID и дата)
- сохранение полученных свойств телефона в файле *\_log/C<номер>.txt*
- сборка apk через `gradlew -PserialKey=<> -PserialValue=<> assembleRelease`
- определение версии приложения (versionName) через парсинг *app/build.gradle*
- переименование apk в Center-<номер>\_<версия>.apk
- установка apk на телефон через `adb install`
- копирование apk на телефон в папку */sdcard/Download/* через `adb push`

Определение ANDROID_ID:
- установка CenterId-Droid на телефон через `adb install`
- неблокирующий запуск сборки логов с телефона через `adb logcat`
- запуск CenterId-Droid через `adb am start` с ожиданием старта активити
- остановка CenterId-Droid через `adb am force-stop`
- остановка сборки логов
- поиск ANDROID_ID в логах
- удаление CenterId-Droid с телефона через `adb uninstall`
