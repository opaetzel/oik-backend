# "Objekte im Kreuzverhör" e-learning Anwendung - Backend

## System-Admin Dokumentation

"Objekte im Kreuzverhör" ist ein Inhaltsverwaltungssystem zugeschnitten auf die Erstellung von Lernmaterialien für Objektwissenschaften.  
Das Backend (dieses Projekt) ist in der Programmiersprache [Go](https://golang.org/) realisiert, das [Frontend](https://github.com/opaetzel/oik-html-frontend) in [emberjs](http://emberjs.com). 

### Download/kompilieren

Um das backend kompilieren zu können, wird eine funktionierende Go-Entwicklungsumgebung mitsamt GOPATH benötigt. Wenn Go korrekt installiert ist, kann das Backend mittels
```bash
go get github.com/opaetzel/oik-backend
```
heruntergeladen und kompiliert werden. Die fertige binary liegt dann in `$GOPATH/bin/oik-backend`.

### Konfiguration

Zum Betrieb wird noch eine Konfiguration benötigt. Eine Beispiel-Konfiguration liegt im repository mit dem Namen [config.toml](./config.toml). Alle Felder kurz erklärt:  

| Key                 | Erklärung                                                           | Mögliche Werte | Default |
|---------------------|---------------------------------------------------------------------|----------------|---------|
| UseTLS              | wenn true -> https  anbieten                                        | true/false     | false   |
| HTTPPort            | Port auf dem HTTP angeboten wird                                    | integer        | 0       |
| HTTPSPort           | Port auf dem HTTPS angeboten wird                                   | integer        | 0       |
| PemFile             | .pem Datei des TLS Zertifikats (wenn UseTLS==true)                  | string         | ""      |
| KeyFile             | .key Datei des TLS Zertifikats (wenn UseTLS==true)                  | string         | ""      |
| DBName              | Name der zu nutzenden Datenbank                                     | string         | ""      |
| DBUser              | Nutzername, mit dem sich in der DB angemeldet wird                  | string         | ""      |
| DBPassword          | Passwort mit dem sich in der DB angemeldet wird                     | string         | ""      |
| ImageStorage        | Ordner, in dem die von Usern hochgeladenen Bilder gespeicert werden | string         | ""      |
| StaticFolder        | Ordner in dem das frontend liegt                                    | string         | ""      |
| AppUrl              | UTL unter der die Applikation von außen erreicht wird               | string         | ""      |
| LogFile             | Pfad des log-files                                                  | string         | ""      |
| MailConfig          | Konfiguration für das Senden der Registrierungs-Mails               | complex        |         |
| MailConfig/UserName | Username mit dem sich am Mailserver angemeldet wird                 | string         | ""      |
| MailConfig/Password | Passwort für den Mailserver                                         | string         | ""      |
| MailConfig/Host     | Host des Mailservers                                                | string         | ""      |
| MailConfig/Port     | Port des smtp-Servers (Mailservers)                                 | integer        | ""      |
| MailConfig/From     | Absenderadresse der gesendeten Mails                                | string         | ""      |

### Datenbank

Als Datenbank wird PostgreSQL ab Version 9.5 vorrausgesetzt. Die in der Konfiguration gesetzte Datenbank muss existieren, der konfigurierte Nutzer ebenso und der Nutzer muss Lese/Schreibzugriff auf die Datenbank haben. Die nötigen Tabellen werden beim Start der Anwendung automatisch angelegt.

## Entwickler Dokumentation

Um die Anwendung zu kompilieren und zum laufen zu bringen, bitte den Schritten oben folgen.  

### Aufbau der Anwendung

Das Backend ist als REST-Service realisiert. Die Routen sind in [routes.go](./routes.go) definiert. Routen haben einen Namen, eine HTTP Methode, die sie akzeptieren, ein Pattern auf das gematcht wird und einen Handler. Alle Handler sind in der Datei [handlers.go](./handlers.go) definiert.

Es gibt drei Arten von Routen:  
- __adminRoutes__: hier wird *vor* dem Ausführen des handlers überprüft, ob der User admin ist. Wenn er es nicht ist, wird mit Status `401` geantwortet
- __editorRoutes__: hier wird *vor* dem Ausführen des handlers überprüft, ob der User editor ist. Wenn er es nicht ist, wird mit Status `401` geantwortet
- __authRoutes__: hier wird *vor* dem Ausführen des handlers überprüft, ob ein valides JWT (hierzu später mehr) im Header gesetzt ist. Wenn es nicht gesetzt ist, wird mit Status `401` geantwortet
- __publicRoutes__: Dies Routen landen auf jeden Fall beim Handler

In der Methode `NewRouter` in der Datei [router.go](./router.go) werden alle Routen zu einem *mux.Router hinzugefügt. Die Routen sind dann von außerhalb über das prefix `api/` gefolgt vom Pattern der jeweiligen Route erreichbar.

### Autentifizierung

Die Autentifizierung mit dem Backend funktioniert über [JWTs](https://jwt.io). Alle Methoden und structs, die etwas mit der Authentifizierung zu tun haben, liegen in [auth.go](./auth.go).

### Modell/Datenbank 

Das Modell ist in der Datei [model.go](./model.go) definiert. Die Datenbankanbindung mitsamt der `CREATE` statements liegt in [db.go](./db.go). Die Interaktion mit der Datenbank funkioniert über selbstgeschriebenes SQL und manuelles Column-Parsing. 
