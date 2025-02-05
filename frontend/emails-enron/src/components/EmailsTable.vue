<template>
  <v-row class="pa-6 flex flex-wrap gap-2 justify-center">
    <v-col
      cols="12"
      lg="6"
      class="rounded-lg shadow-md bg-white p-4 border-2 border-cyan-100 rounded-lg"
    >
      <h4 class="text-h5 font-weight-medium text-disabled">
        Registros Enron Emails en Zinc Search
      </h4>
      <v-card title="" elevation="0" class="overflow-hidden">
        <template v-slot:text>
          <v-text-field
            v-model="search"
            label="Buscar en el 'Body' del correo"
            variant="outlined"
            hide-details
            single-line
            @keydown.enter="handleSearch"
            class="icon-blue"
          >
            <template v-slot:prepend-inner>
              <SearchIcon stroke-width="1.5" size="17" class="text-lightText SearchIcon" />
            </template>
          </v-text-field>
        </template>

        <v-data-table
          class="p-4"
          @click:row="selectRow"
          :headers="headers"
          :items="items"
          :loading="loading"
          v-model:items-per-page="itemsPerPage"
          hide-default-footer
        >
        </v-data-table>
        <div class="bg-indigo-100 pa-2 rounded-lg shadow-md mt-2">
          <span class="text-sm">Registros: {{ paginationText }}</span>
        </div>
        <div class="grid place-items-center pb-5 pt-5">
          <div class="pt-2 flex flex-wrap gap-2 justify-center">
            <v-btn color="primary" @click="goToFirstPage" :disabled="page === 1">Primera</v-btn>
            <v-btn color="primary" @click="goToPreviousPage" :disabled="page === 1">Anterior</v-btn>
            <span>Página {{ page }} de {{ totalPages }}</span>
            <v-btn color="primary" @click="goToNextPage" :disabled="page === totalPages"
              >Siguiente</v-btn
            >
            <v-btn color="primary" @click="goToLastPage" :disabled="page === totalPages"
              >Última</v-btn
            >
          </div>
        </div>

        <!-- PAGINACIÓN MANUAL
    <div class="pagination">
      <v-btn @click="goToFirstPage" :disabled="page === 1">Primera</v-btn>
      <v-btn @click="goToPreviousPage" :disabled="page === 1">Anterior</v-btn>
      <span>Página {{ page }} de {{ totalPages }}</span>
      <v-btn @click="goToNextPage" :disabled="page === totalPages">Siguiente</v-btn>
      <v-btn @click="goToLastPage" :disabled="page === totalPages">Última</v-btn>
    </div>-->
      </v-card>
    </v-col>
    <v-col
      cols="12"
      lg="5"
      class="rounded-lg shadow-md bg-white p-4 border-2 border-cyan-100 rounded-lg"
      v-if="selectedItem"
    >
      <div class="d-flex align-center ga-4">
        <div>
          <h4 class="text-h5 font-weight-medium text-disabled pb-4">Detalles del correo</h4>

          <v-row>
            <v-col cols="12" xs="12" sm="6" md="6">
              <p><strong>ID:</strong></p>
            </v-col>
            <v-col cols="12" xs="12" sm="6" md="6">
              <p>{{ selectedItem.id }}</p>
            </v-col>
            <v-col cols="12" xs="12" sm="6" md="6">
              <p><strong>Subject:</strong></p>
            </v-col>
            <v-col cols="12" xs="12" sm="6" md="6">
              <p>{{ selectedItem.subject }}</p>
            </v-col>
            <v-col cols="12" xs="12" sm="6" md="6">
              <p><strong>From:</strong></p>
            </v-col>
            <v-col cols="12" xs="12" sm="6" md="6">
              <p>{{ selectedItem.from }}</p>
            </v-col>
            <v-col cols="12" xs="12" sm="6" md="6">
              <p><strong>To:</strong></p>
            </v-col>
            <v-col cols="12" xs="12" sm="6" md="6">
              <p>{{ selectedItem.to }}</p>
            </v-col>
            <v-col cols="12" xs="12" sm="6" md="6">
              <p><strong>Date:</strong></p>
            </v-col>
            <v-col cols="12" xs="12" sm="6" md="6">
              <p>{{ selectedItem.date }}</p>
            </v-col>
            <v-col cols="12" lg="12">
              <p><strong>Body:</strong></p>
              <p v-html="formattedBody(selectedItem.body)"></p>
            </v-col>
          </v-row>
        </div>
      </div>
      <!--<v-card
        color="info"
        elevation="0"
        class="overflow-hidden pa-4 text-center"
        v-if="selectedItem"
      >
        <v-row>
          <v-col cols="12" lg="12" v-if="selectedItem">
            <v-card elevation="0" class="bubble-shape-sm overflow-hidden bubble-warning">
              <v-card-text class="pa-5">
                <div class="d-flex align-center ga-4">
                  <div>
                    <v-row>
                      <v-col cols="12" lg="12" class="justified-text">
                        <p><strong>Body:</strong></p>
                        <p v-html="formattedBody(selectedItem.body)"></p>
                      </v-col>
                    </v-row>

                    <span class="text-subtitle-2 text-disabled font-weight-medium"
                      >Total Income</span
                    >
                  </div>
                </div>
              </v-card-text>
            </v-card>
          </v-col>
        </v-row>
      </v-card>-->
    </v-col>
  </v-row>
</template>

<script>
import axios from 'axios'

export default {
  name: 'EmailsTable',

  data() {
    return {
      search: '',
      search2: '',
      page: 1,
      itemsPerPage: 10,
      totalItems: 0,
      loading: false,
      selectedItem: null,
      headers: [
        { key: 'id', title: 'ID' },
        { key: 'from', title: 'From' },
        { key: 'to', title: 'To' },
        { key: 'subject', title: 'Subject' },
      ],
      items: [],
    }
  },
  computed: {
    totalPages() {
      return Math.ceil(this.totalItems / this.itemsPerPage) || 1
    },
    paginationText() {
      if (this.totalItems === 0) {
        return 'No hay resultados.'
      }
      const start = (this.page - 1) * this.itemsPerPage + 1
      const end = Math.min(start + this.itemsPerPage - 1, this.totalItems)
      return `${start}-${end} de ${this.totalItems}`
    },
  },
  methods: {
    async fetchData() {
      this.loading = true
      try {
        const from = (this.page - 1) * this.itemsPerPage
        const url = `http://localhost:8090/api/documents?index=emails2&from=${from}&size=${this.itemsPerPage}&search=${this.search}`
        const response = await axios.get(url)

        if (response.request.status == 200) {
          this.totalItems = response.data.hits.total.value
          this.items = response.data.hits.hits.map((hit) => ({
            id: hit._source.id,
            from: hit._source.from,
            to: hit._source.to,
            subject: hit._source.subject || 'No Subject',
            date: hit._source.date,
            body: hit._source.body,
          }))
        } else {
          this.totalItems = 0
          this.items = []
        }
      } catch (error) {
        console.error('Error fetching data:', error)
      } finally {
        this.loading = false
      }
    },

    handleSearch() {
      this.page = 1
      this.fetchData()
    },

    goToFirstPage() {
      this.page = 1
      this.fetchData()
    },
    goToPreviousPage() {
      if (this.page > 1) {
        this.page--
        this.fetchData()
      }
    },
    goToNextPage() {
      if (this.page < this.totalPages) {
        this.page++
        this.fetchData()
      }
    },
    goToLastPage() {
      this.page = this.totalPages
      this.fetchData()
    },
    selectRow(event, { item }) {
      console.log('Fila seleccionada: ', item) // Depuración
      this.selectedItem = item // Asignar datos de la fila seleccionada
    },
    formattedBody(body) {
      // Reemplazar saltos de línea por <br> para HTML
      let formattedText = body.replace(/\n/g, '<br/>')

      // Buscar y envolver el texto en negrita

      if (this.search != '') {
        const escapedSearch = this.search.replace(/[.*+?^=!:${}()|[\]/\\]/g, '\\$&')
        const regex = new RegExp(`(${escapedSearch})`, 'gi') // Buscar con insensibilidad a mayúsculas

        formattedText = formattedText.replace(regex, '<strong>$1</strong>')
      }

      return formattedText
    },
  },
  mounted() {
    this.fetchData()
  },
}
</script>
