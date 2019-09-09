FROM jjanzic/docker-python3-opencv:latest
WORKDIR ./
COPY requirements.txt ./
RUN pip3 install --no-cache-dir -r requirements.txt
COPY . .
CMD ["python3", "-u", "./check.py"]