# binger
Download de imagens com base em termos de busca fazendo uso de Image Search API v7, da Microsoft Azure

```
~# binger -q "<termos_de_busca>"
```

__binger__ cria uma pasta chamada `files` na pasta corrente e salva nela as imagens encontradas.
Você também pode especificar uma pasta já existente no computador:

```
~# binger -q "<termos_de_busca>" -d "<caminho_pra_sua_pasta_existente>"
```

__binger__ pode não encontrar todas as imagens. Para acelerar o processo, cada imagem tem um timeout de 20 segundos.

Este código contém uma chave padrão da __Image Search API v7__. Não existem garantias de que ela continuará
funcionando.
Para continuar utilizando __binger__, você deve especificar sua própria chave e utilizá-la assim:

```
~# binger -q "<termos_de_busca>" -k <sua_chave>
```

Para criar sua chave, acesse 
https://docs.microsoft.com/pt-br/rest/api/cognitiveservices/bing-images-api-v7-reference.