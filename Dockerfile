FROM zenika/alpine-chrome
COPY fonts/*.ttf /usr/share/fonts/
RUN fc-cache -fv
COPY bin/proxy .

EXPOSE 9222
ENTRYPOINT ["./proxy"]
