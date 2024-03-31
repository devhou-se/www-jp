FROM python:3.12
WORKDIR /app
COPY ./requirements.txt ./requirements.txt
RUN pip install -r requirements.txt
COPY ./reactions.py ./reactions.py
RUN chmod +x ./reactions.py

ENTRYPOINT ["python", "./reactions.py"]

