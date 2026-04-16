FROM python:3.11-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

RUN mkdir -p /app/sessions /app/data

# По умолчанию запускаем TG-воркер.
# VK-воркер переопределяет CMD через docker-compose: command: python vk_monitor.py
CMD ["python", "tg_monitor.py"]
