# Checklist obrigatório antes de alterar frontend ou API

Este checklist deve ser seguido em toda mudança que afete o contrato, o backend ou a interface.

1. Atualizar o contrato canônico em api-contract.json.
2. Validar nomes de campo, tipos, endpoints e status esperados.
3. Confirmar compatibilidade com o backend em main.go.
4. Confirmar compatibilidade com o frontend em static/.
5. Rodar a suíte Go e validar a sintaxe do frontend antes de publicar.
6. Registrar o impacto em testes, filtros, resultados e ignorados.
7. Abrir um PR com resumo da mudança, risco e evidência de validação.
8. Aplicar os labels `contract-reviewed` e `compatibility-reviewed` após revisar os respectivos itens.
9. Obter ao menos uma aprovação humana antes do merge.
10. Só então liberar a alteração.

Processo formal de PR:
- O PR deve explicitar: escopo, contrato afetado, testes executados e resultado esperado.
- Se o contrato mudar, o mesmo ajuste deve aparecer em código, testes e documentação.
- A aprovação formal deve ocorrer antes do merge final.

Regras:
- Nenhuma mudança de frontend ou API pode entrar sem passar por este checklist.
- Nenhuma mudança de contrato pode ser publicada sem PR revisado e validado.
