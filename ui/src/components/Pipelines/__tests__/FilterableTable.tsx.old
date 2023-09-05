import { fireEvent } from '@testing-library/react';

// WIP - Make a sharable class to test all Filterable table functionality

export class TestFilterableTable {
  constructor(_tableId: string) {
    this.tableId = _tableId;
    this.fireEvent = fireEvent;
  }
  tableId: string = '';
  fireEvent: any;

  getTableInfo() {
    const tbl = document.querySelector(`#${this.tableId} table`);
    const rows = tbl?.querySelectorAll('tbody tr');
    const headers = tbl?.querySelectorAll('thead tr th');
    return { rows, headers };
  }
  getRowInfoByIndex(rowIndex: number) {
    const rows = document.querySelectorAll(`#${this.tableId} tbody tr`);
    return rows[rowIndex].querySelectorAll('td');
  }

  sortTableByColumn(columnName: string) {
    const btns = document.querySelectorAll<HTMLElement>(
      `#${this.tableId} table thead tr th button`,
    );
    btns.forEach(ele => {
      if (ele.textContent === columnName) {
        ele.click();
      }
    });
  }
  searchTableByValue(searchVal: string) {
    const searchBtn = document.querySelector<HTMLElement>(
      `#${this.tableId} button[class*='SearchField']`,
    );
    searchBtn?.click();
    const searchInput = document.getElementById(
      'table-search',
    ) as HTMLInputElement;

    this.fireEvent.change(searchInput, { target: { value: searchVal } });

    const searchForm = document.querySelector(
      `#${this.tableId} div[class*='SearchField'] > form`,
    ) as Element;

    this.fireEvent.submit(searchForm);
    return this.getTableInfo();
  }

  applyFilterByValue(filterIndex: number, value: string) {
    const filterBtn = document.querySelector<HTMLElement>(
      `#${this.tableId} button[class*='FilterableTable']`,
    );
    filterBtn?.click();

    const filters = document.querySelectorAll<HTMLElement>(
      `#${this.tableId} form > ul > li`,
    );
    const filterInput = filters[filterIndex].querySelector<HTMLElement>(
      `input[id="${value}"]`,
    );
    filterInput?.click();
    return this.getTableInfo();
  }

  testRowValues = (rowValue: NodeListOf<Element>, matches: Array<string>) => {
    for (let index = 0; index < rowValue.length; index++) {
      const element = rowValue[index];
      expect(element.textContent).toEqual(matches[index]);
    }
  };

  testRenderTable(displayedHeaders: Array<string>, rowLength: number) {
    const { rows, headers } = this.getTableInfo();
    expect(headers).toHaveLength(displayedHeaders.length);
    expect(rows).toHaveLength(rowLength);
    this.testRowValues(headers!, displayedHeaders);
  }

  testSearchTableByValue(
    searchValue: string,
    targetRowIndex: number,
    rowValues: Array<string>,
  ) {
    const { rows } = this.searchTableByValue(searchValue);
    expect(rows).toHaveLength(1);
    const tds = rows![targetRowIndex].querySelectorAll('td');
    this.testRowValues(tds, rowValues);
  }

  testFilterTableByValue(
    filterIndex: number,
    value: string,
    rowValues: Array<string>,
  ) {
    const { rows } = this.applyFilterByValue(filterIndex, value);

    expect(rows).toHaveLength(1);
    const tds = rows![0].querySelectorAll('td');

    this.testRowValues(tds, rowValues);
  }
}
