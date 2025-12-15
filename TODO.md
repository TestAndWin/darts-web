- Anzahl Legs pro Set >> Als Option beim Spielstart
- readme.md aktualisieren
- TODO.md und Readme aktualisieren und zusammenfassen

## Installation 
### Minikube auf Ubunto 
```bash
minikube start --driver=docker
minikube addons enable metrics-server
minikube config view
minikube addons enable ingress

sudo apt-get update
sudo apt-get install -y nginx
```

### nginx konfigurieren
```bash
sudo nano /etc/nginx/sites-available/darts-proxy
```

```nginx
  server {
      listen 80;
      server_name mini-pc;

      location / {
          proxy_pass http://192.168.49.2:32622;
          proxy_set_header Host mini-pc;
          proxy_set_header X-Real-IP $remote_addr;
          proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
          proxy_set_header X-Forwarded-Proto $scheme;
      }
  }
```

## Zusammenfassung der Lösung

Problem: Minikube läuft in einer VM und die internen Cluster-IPs sind von außen nicht erreichbar.

Lösung: nginx als Reverse Proxy auf dem Host (mini-pc), der zu Minikube's NodePort weiterleitet.

Finale Architektur:

Browser (http://mini-pc/darts/)
    ↓
nginx auf mini-pc Port 80
    ↓ (proxy_pass mit Host: mini-pc)
Minikube Ingress Controller (192.168.49.2:32622)
    ↓
Kubernetes Service (darts-darts-web:80)
    ↓
Pod (darts-app:8080)

Wichtige Konfigurationen:

- nginx: /etc/nginx/sites-available/darts-proxy - leitet alle Requests mit Host: mini-pc Header weiter
- Kubernetes Ingress: Matched auf host: mini-pc und path: /darts
- NodePort: 32622 (HTTP) und 31828 (HTTPS)

## Daten Persistenz
  Die Daten bleiben erhalten, weil Sie PersistentVolumeClaim (PVC) verwenden.

  Wie es funktioniert

  Bei einem Update/Deployment

  ### Neue Version deployen
  ```bash
  helm upgrade darts-web ./charts/darts-web
  ```

  Was passiert:
  1. ✅ Alter Pod wird beendet
  2. ✅ Neuer Pod mit neuem Image wird gestartet
  3. ✅ PVC bleibt erhalten und wird wieder gemountet
  4. ✅ /data/darts.db ist wieder verfügbar mit allen Daten

  Lifecycle-Unterschied

  | Ressource               | Lebensdauer                                |
  |-------------------------|--------------------------------------------|
  | Pod                     | Vergänglich - wird bei Update neu erstellt |
  | PersistentVolumeClaim   | Persistent - bleibt erhalten               |
  | Daten in /data/darts.db | Persistent - im PVC gespeichert            |

  Wann würden Daten verloren gehen?

  Nur wenn Sie explizit das PVC löschen:

  # ⚠️ DAS würde Daten löschen:
  kubectl delete pvc darts-pvc -n darts

  # oder bei Helm (je nach --keep-history Flag):
  helm uninstall darts-web
  kubectl delete pvc darts-web-... # PVC muss separat gelöscht werden

  Best Practice: Backups

  Obwohl die Daten persistent sind, sollten Sie trotzdem Backups machen:

  # Datenbank aus dem Pod kopieren
  kubectl cp darts/darts-app-xxxxx:/data/darts.db ./backup-darts.db

  # Oder in den Pod gehen und manuell kopieren
  kubectl exec -it -n darts darts-app-xxxxx -- cp /data/darts.db /tmp/

  Fazit: Ihre Konfiguration ist richtig - die Datenbank überlebt alle Updates! 👍

## Deployment
Image-Transfer via TAR-Datei (ohne Registry)

  # Image bauen und exportieren
  make export

1. scp darts-app-${VERSION}.tar <user>@mini-pc:/tmp/
2. ssh <user>@mini-pc
3. minikube image load /tmp/darts-app-${VERSION}.tar
4. Update charts/darts-web/values.yaml with tag: "${VERSION}"
5. helm upgrade darts ./charts/darts-web